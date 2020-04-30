package cmd

import "gitlab.com/youtopia.earth/ops/snip/config"

func newBashCompletionFunc(cl *config.ConfigLoader) string {
	var bashCompletionFunc = `
	__snip_get_namespace(){
    local namespaceflag=$(echo "${words[@]}" | sed -E 's/.*(--namespace[= ]+|-s +)([A-Za-z_]+).*/\2/')
  if [[ $namespaceflag = *" "* ]]; then
      namespaceflag="$` + cl.PrefixEnv("NAMESPACE") + `"
  fi
	if [ "$namespaceflag" = "" ]; then
      namespaceflag=$(snip config namespace)
  fi
  echo "$namespaceflag"
  }
  __snip_custom_func() {
    case ${last_command} in
        snip_clean_secrets|snip_rotate-random)
          __snip_get_secret_root_names
          return
          ;;
        *)
        ;;
    esac
  }`
	return bashCompletionFunc
}
