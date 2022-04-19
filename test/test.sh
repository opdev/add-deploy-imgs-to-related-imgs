#!/usr/bin/env bash

#
# - Copy the fixture into a temporary location
# - Modify it in place
# - Check to ensure the replaced file has 
#   the expected image references.
# 

this_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
fixtures_dir="${this_dir}/fixtures"
root_dir="${this_dir}/.."
bin_name="add-deploy-imgs-to-related-imgs"
mod_tool="${root_dir}/${bin_name}"
csv_fixtures="clusterserviceversion-*.yaml"

# Check for requirements. Note that `yq` required here is this one
# https://github.com/mikefarah/yq/ - at least v 4.23.1
echo "Checking system for requirements."
requirements="find mktemp stat yq"
for r in $requirements ; do
    which $r &>/dev/null \
        || { 
            echo "ERR Could not find requirement \"${r}\" in path." \
            && exit 1
        }
done

echo "Verifying the test binary is built and lives at a known path."
# Check that the binary exists in the root directory.
stat "${root_dir}/${bin_name}" &>/dev/null \
    || {
        echo "ERR Could not find the \"${bin_name}\" in the root directory"
        echo "    Make sure to build the binary before running the tests!" \
        && exit 5
        }

# Copy the test csvs into a temporary location so we don't
# clutter up our repository.
echo "Creating a temporary directory and copying test fixtures into it."
tempdir=$(mktemp -d)
find "${fixtures_dir}" -type f -name 'clusterserviceversion-*.yaml' -exec cp -a {} "${tempdir}" \;

# Run the tool against each fixture and parse results.
failed_tests=0

echo "Test: the tool must successfully add related images when the key is missing from the CSV."
file="${tempdir}/clusterserviceversion-missing-key.yaml"
"${mod_tool}" "${file}"
len=$(yq .spec.relatedImages ${file} | yq '. | length'); test $len -eq 2  \
    || {
        echo "FAILED Something happened while trying to add related images when the key \"relatedImages\" was missing in the CSV."
        ((failed_tests=failed_tests+1))
        # Don't exit here
       }
unset file len

echo "Test: the tool must add related images even if the key exists, but has an empty array value."
file="${tempdir}/clusterserviceversion-with-empty.yaml"
"${mod_tool}" "${file}"
len=$(yq .spec.relatedImages ${file} | yq '. | length'); test $len -eq 2 \
    || {
        echo "FAILED Something happened while trying to add related images when the key \"relatedImages\" was set to \"[]\"."
        ((failed_tests=failed_tests+1))
        # Don't exit here
       }
unset file len

echo "Test: the tool must add related images even if preexisting unique images also exist."
file="${tempdir}/clusterserviceversion-with-preexisting.yaml"
"${mod_tool}" "${file}"
len=$(yq .spec.relatedImages ${file} | yq '. | length'); test $len -eq 3 \
    || {
        echo "FAILED Something happened while trying to add additional related images with preexisting related images."
        ((failed_tests=failed_tests+1))
        # Don't exit here
       }
unset file len

# Evaluate if we failed any tests.
if [ "${failed_tests}" != "0" ]; then 
    echo "Some test failed! The temp directory was:"
    echo "   ${temp_dir}."
    echo "Exiting."
    exit 15
fi

echo "All tests passed!"
