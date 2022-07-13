#!/usr/bin/env bash

package="github.com/AutomatedProcessImprovement/waiting-time-backend"
package_name="waiting-time-backend"

platforms=("linux/amd64")

for platform in "${platforms[@]}"
do
	platform_split=(${platform//\// })
	GOOS=${platform_split[0]}
	GOARCH=${platform_split[1]}
	dir_name="$GOOS-$GOARCH"
	build_dir="build/$dir_name"
	output_name="$build_dir/$package_name"
	if [ $GOOS = "windows" ]; then
		output_name+='.exe'
	fi

	# providing static assets for the server
	mkdir -p "$build_dir/assets"
	cp -r assets/samples "$build_dir/assets"

	env GOOS="$GOOS" GOARCH="$GOARCH" go build -o $output_name $package
	if [ $? -ne 0 ]; then
   		echo 'An error has occurred! Aborting the script execution...'
		exit 1
	fi

	# archive the builds
	(
    cd build
    if [ $GOOS = "windows" ]; then
      zip -r "$dir_name.zip" "$dir_name"
    else
      tar -czf "$dir_name.tar.gz" "$dir_name"
    fi
  )

  # remove the build folder
  rm -rf "$build_dir"
done