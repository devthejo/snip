#!/usr/bin/env bash
set -e

SNIP_VERSION=${SNIP_VERSION:-latest}
SNIP_GITLAB_PROJECT_ID=`curl -s https://gitlab.com/api/v4/projects/ytopia%2Fops%2Fsnip/ | jq .id`
[ "$SNIP_VERSION" = "latest" ] && SNIP_VERSION=`curl -s "https://gitlab.com/api/v4/projects/${SNIP_GITLAB_PROJECT_ID}/repository/tags?search=^v&order_by=updated&sort=desc" | jq -r '.[0] | .name'`
SNIP_ARTIFACT_URL=`curl -s https://gitlab.com/api/v4/projects/$SNIP_GITLAB_PROJECT_ID/releases/${SNIP_VERSION}/assets/links | jq -r '.[] | select(.name | contains("snip_'${SNIP_VERSION}'_linux_amd64")) | .url'`
curl -o snip -s $SNIP_ARTIFACT_URL
chmod +x snip
mv snip /usr/local/bin/
which snip

echo installed