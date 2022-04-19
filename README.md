# Add Deployment Images to Related Images

A simple tool that adds your ClusterServiceVersion's ("CSV") deployment images to the
`.spec.relatedImages` section.

This tool currently handles:

- a CSV that does not have a relatedImages key at all.
- a CSV that has an empty relatedImages key (e.g. `relatedImages: []`).
- a CSV with pre-existing relatedImages.

Note the following behaviors:

- Your container image's name is what's used for the `RelatedImage.Name` field.
- RelatedImage values must be images referenced via digest. Please pin your images before running this tool ([get a pin tool](https://github.com/opdev/pin-deply-imgs-in-csv)). This tool does not validate that this has taken place before adding the entries.
- This will modify your ClusterServiceVersion directly. Please use source control prior to running this tool so that you can return to a previous state (should we inadvertently mangle your CSV).

## Usage

```
add-deploy-imgs-to-related-imgs /path/to/clusterserviceversion.yaml

This tool will take all container images in your
ClusterServiceVersion's deployment(s) and add them to the relatedImages
field. This tool expects that your deployment's tools have already been
pinned to a digest.
```

To see the help output (same as above), run `add-deploy-imgs-to-related-imgs help`.

To see the version, run `add-deploy-imgs-to-related-imgs version`.