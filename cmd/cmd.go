package cmd

import (
	"fmt"
	"os"
	"strings"

	dockerparser "github.com/novln/docker-parser"
	"github.com/opdev/add-deploy-imgs-to-related-imgs/replacers"

	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"sigs.k8s.io/yaml"
)

var (
	usage = `This tool will take all container images in your
ClusterServiceVersion's deployment(s) and add them to the relatedImages
field. This tool expects that your deployment's tools have already been
pinned to a digest.`
	execution = fmt.Sprintf("%s /path/to/clusterserviceversion.yaml", os.Args[0])
	version   = "unknown" // replace with -ldflags at build time.
)

func Run() int {
	// Expect only a single positional arg - the csv file path
	if len(os.Args) != 2 {
		fmt.Fprintf(
			os.Stderr,
			"ERR accepts only a single positional arg: the path"+
				"to the CSV file to modify. Received %d",
			len(os.Args)-1)
		return 10
	}

	// Handle requests for help or the version
	switch strings.ToLower(os.Args[1]) {
	case "help":
		printUsage()
		return 0
	case "version":
		fmt.Println(version)
		return 0
	}

	// Read in CSV file
	csvFile := os.Args[1]
	bts, err := os.ReadFile(csvFile)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Unable to read the file on disk at path %s with error\n%s",
			csvFile,
			err,
		)
		return 20
	}

	var csv operatorsv1alpha1.ClusterServiceVersion

	// Decode data as  yaml
	err = yaml.Unmarshal(bts, &csv)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"The CSV file provided %s did not cleanly marshal to a CSV struct with error\n%s",
			csvFile,
			err,
		)
		return 30
	}

	// Grab the container images in the deployment spec and add them to the related images struct.
	containersInCSV := map[string]string{}
	for _, deploymentSpec := range csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs {
		for _, container := range deploymentSpec.Spec.Template.Spec.Containers {
			// TODO: handle this error.
			imageRef, _ := dockerparser.Parse(container.Image)

			// shortName here includes the repository, but we only want the image
			// name itself. Try to split the repository from the image name, and
			// if we can't split into two values, then we assume the shortName
			// is just the image name.
			shortName := imageRef.ShortName()
			split := strings.Split(shortName, "/")
			imageKey := shortName
			if len(split) > 1 {
				imageKey = split[1]
			}
			containersInCSV[imageKey] = container.Image
			// TODO: Does it make sense to expand fat manifests and add them
			// to related images? What value would we store in RelatedImage.Name?
		}
	}

	for name, image := range containersInCSV {
		csv.Spec.RelatedImages = append(csv.Spec.RelatedImages, operatorsv1alpha1.RelatedImage{
			Name:  name,
			Image: image,
		})
	}

	replacer := replacers.InPlaceRelatedImagesReplacer{
		OriginalCSVBytes: bts,
		NewRelatedImages: csv.Spec.RelatedImages,
	}

	newCSVBytes, err := replacer.Replace()
	if err != nil {
		fmt.Fprintf(os.Stderr, "there was an error replacing the related images in memory:\n%s", err)
		return 40
	}

	// write the file in-place
	err = os.WriteFile(csvFile, newCSVBytes, 0644)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Unable to overwrite the file at path %s with error\n%s",
			csvFile,
			err,
		)

		// TODO: should we consider sending the newCSV to stdout?
		return 70
	}

	return 0

}

func printUsage() {
	fmt.Printf(`%s
	
%s`, execution, usage,
	)
}
