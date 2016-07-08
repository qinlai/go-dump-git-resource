#!/bin/sh
tar -zxf git_projects.tar.gz
run_path=`pwd`
project='project1'
../bin/dump_git_resource_for_mac -project=$project -base_tag=201601 -target_dir=$run_path/dump -git_dir=$run_path/git_projects
../bin/dump_git_resource_for_mac -project=$project -base_tag=201601 -end_tag=201602 -target_dir=$run_path/dump -git_dir=$run_path/git_projects
