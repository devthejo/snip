#!/bin/sh

progname=$1

for os in ${OSS}
do
	fullname="${progname}_${CI_COMMIT_TAG}_${os}_${GOARCH}"
	# linkname="${fullname}\ (SHA256 $(cut -f1 -d' ' ${fullname}.sha256))"
	linkname="${fullname}"

	linkurl="${BASEURL}/jobs/${CI_JOB_ID}/artifacts/${fullname}"

	# curl --request POST --header "PRIVATE-TOKEN: ${PRIVATE_TOKEN}" --form "file=@$fullname" "https://gitlab.com/api/v4/projects/$CI_PROJECT_ID/uploads"
	linklist="${linklist}{\"name\": \"${linkname}\", \"url\": \"${linkurl}\"}"
done
links="[$(echo ${linklist}|sed 's/}{/}, {/g')]"

descr="$(curl -H \"PRIVATE-TOKEN:\ ${PRIVATE_TOKEN}\" ${BASEURL}/repository/tags/${CI_COMMIT_TAG}|jq -r '.message')"

DATA="
{
  \"name\": \"${progname} version ${CI_COMMIT_TAG}\",
  \"description\": \"${descr}\",
  \"tag_name\": \"${CI_COMMIT_TAG}\",
  \"assets\": {
    \"links\": "${links}"
  }
}
"
curl -H 'Content-Type: application/json' -X POST -H "PRIVATE-TOKEN: ${PRIVATE_TOKEN}" "${BASEURL}/releases" -d "${DATA}"
