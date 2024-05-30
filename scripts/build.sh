repo_root_dir=$(pwd)
cmd=./cmd/server
target=./bin/server

echo "env is $env"
echo "ENV is $ENV"

# use goc in test env
if [ "$env"X = "test"X ]; then
    echo "download goc for scp"
    curl -o goc 'https://goc.epd.i.test.shopee.io/goc/api/v2/goc_binary_file/download'
    chmod u+x goc
    echo "curr dir is $(pwd)"

    echo "cd to $cmd"
    cd $cmd

    echo "goc build"
    $repo_root_dir/goc build -o $repo_root_dir/$target . --debug

    echo "cd to $repo_root_dir"
    cd $repo_root_dir
else
    spkit build .
    ls -lah -d $target
fi