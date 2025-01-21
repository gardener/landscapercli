#!/usr/bin/env python3

import pathlib
from ci.util import (
    check_env,
    ctx,
)
import ccc.github

from github.util import GitHubRepositoryHelper

OUTPUT_FILE_NAME='build-result'
VERSION_FILE_NAME='VERSION'

repo_owner_and_name = check_env('SOURCE_GITHUB_REPO_OWNER_AND_NAME')
repo_dir = check_env('MAIN_REPO_DIR')
output_dir = check_env('OUT_PATH')

repo_owner, repo_name = repo_owner_and_name.split('/')

repo_path = pathlib.Path(repo_dir).resolve()
version_file_path = repo_path / VERSION_FILE_NAME

version_file_contents = version_file_path.read_text()

cfg_factory = ctx().cfg_factory()
github_cfg = cfg_factory.github('github_com')
github_api = ccc.github.github_api(github_cfg)

github_repo_helper = GitHubRepositoryHelper(
    owner=repo_owner,
    name=repo_name,
    github_api=github_api,
)

gh_release = github_repo_helper.repository.release_from_tag(version_file_contents)

output_path = pathlib.Path(output_dir).resolve()
output_files = list(output_path.glob('*.gz'))
for output_file in output_files:
    gh_release.upload_asset(
        content_type='application/gzip',
        name=output_file.name,
        asset=output_file.open(mode='rb'),
    )