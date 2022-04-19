package replacers

import (
	"bytes"
	"fmt"
	"regexp"

	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"sigs.k8s.io/yaml"
)

// InPlaceRelatedImagesReplacer facilitate replacing a relatedImages block in an existing CSV
type InPlaceRelatedImagesReplacer struct {
	OriginalCSVBytes []byte
	NewRelatedImages []operatorsv1alpha1.RelatedImage

	newRelatedImagesBytes []byte
	// if - relatedImages exists at all
	missingRelatedImages bool
	// if - relatedImages: {}
	relatedImagesIsEmpty bool
}

// parseExistingRelatedImages determins the values of containsRelatedImagesSection
// and relatedImagesIsEmpty
func (r *InPlaceRelatedImagesReplacer) parseExistingRelatedImages() {
	r.missingRelatedImages = !bytes.Contains(r.OriginalCSVBytes, []byte("relatedImages:"))
	r.relatedImagesIsEmpty = bytes.Contains(r.OriginalCSVBytes, []byte("relatedImages: []"))
}

// renderNewRelatedImages converts NewRelatedImages into a byte slice with the "relatedImages"
// key in palce at the top level. This includes adjusting spacing of the overall byte slice
// to make sure it would render with appropriate indentation. Since the relatedImages key
// is a top level key of spec, we render everything with two spaces of indentation.
func (r *InPlaceRelatedImagesReplacer) renderNewRelatedImages() {
	standardIndent := []byte("  ")
	key := []byte("relatedImages:")

	relatedImagesBytes, _ := yaml.Marshal(r.NewRelatedImages)
	// Marshal does not include the relatedImages key itself, so add it.
	relatedImagesBytes = bytes.Join([][]byte{key, relatedImagesBytes}, []byte("\n"))

	// The array of images is rendered at the top level by Marshal, so we
	// need to fix the indentation of each line.
	split := bytes.Split(relatedImagesBytes, []byte("\n"))
	for i, line := range split {
		if len(line) > 0 {
			split[i] = bytes.Join([][]byte{standardIndent, line}, []byte(""))
		}
	}

	// If we have an extra space or new line at the end, trim the space.
	if len(split[len(split)-1]) == 0 {
		split = split[0 : len(split)-1]
	}

	// Rejoin the entire array into a single byte slice.
	final := bytes.Join(split, []byte("\n"))
	r.newRelatedImagesBytes = final
}

// Replace will replace the relatedImages section of the OriginalCSVBytes
// with NewRelatedImages, and return the new value.
func (r *InPlaceRelatedImagesReplacer) Replace() ([]byte, error) {
	r.parseExistingRelatedImages()
	r.renderNewRelatedImages()

	if r.missingRelatedImages {
		// This regexp looks for the top-level spec: key
		re := regexp.MustCompile(`(?m)^spec:$`)
		final := re.ReplaceAll(r.OriginalCSVBytes, bytes.Join(
			[][]byte{[]byte("spec:"), r.newRelatedImagesBytes}, []byte("\n")))
		return final, nil
	}

	if r.relatedImagesIsEmpty {
		// this Regexp grabs the empty slice with at exactly two leading spaces
		re := regexp.MustCompile(`(?m)\s{2}relatedImages.*$`)
		final := re.ReplaceAll(r.OriginalCSVBytes, r.newRelatedImagesBytes)
		return final, nil
	}

	// replace existing entries with our new values.
	// regexp matches the entire relatedImages block in the first grouping,
	// and the next entry at the same indentation level as relatedImages:
	// in the second group. The latter is discarded.
	re := regexp.MustCompile(`(?msU)(\s\srelatedImages:.+)\n(^\s{2}[a-z])`)
	if !re.Match(r.OriginalCSVBytes) {
		return nil, fmt.Errorf("for some reason, we couldn't match the relatedImages section in your CSV")
	}

	// get the indexes of the match locations. index 2/3 are the indexes of the
	// first subgroup, which is what we want.
	loc := re.FindSubmatchIndex(r.OriginalCSVBytes)

	// outputNewRIB := append([][]byte{[]byte("  relatedImages:")}, splitNewRIB...)

	outputCSV := bytes.Replace(
		r.OriginalCSVBytes,                // in this data
		r.OriginalCSVBytes[loc[2]:loc[3]], // replace this section
		r.newRelatedImagesBytes,           // with this
		1,                                 // exactly one time
	)

	return outputCSV, nil
}
