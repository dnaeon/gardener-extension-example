#!/usr/bin/env bash
# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

#
# Utility script for bootstrapping an extension skeleton

set -e

_SCRIPT_NAME="${0##*/}"
_SCRIPT_DIR=$( dirname "$( readlink -f -- "${0}" )" )
_PROJECT_DIR="${_SCRIPT_DIR}/.."
_TOOLS_MOD_FILE="${_PROJECT_DIR}/internal/tools/go.mod"

# Set to true in order to enable debugging
DEBUG=${DEBUG:-false}

# Set to true in order to preserve Git history
PRESERVE_HISTORY=${PRESERVE_HISTORY:-false}

if [[ ${DEBUG} == "true" ]]; then
  set -x
fi

# Refer to the ANSI escape codes table for more details.
# https://en.wikipedia.org/wiki/ANSI_escape_code
_RED='\033[0;31m'
_GREEN='\033[0;32m'
_NO_COLOR='\033[0m'

# Displays an INFO message
#
# $1: message to display
function _msg_info() {
  local _msg="${1}"

  echo -e "${_SCRIPT_NAME} | ${_GREEN}INFO${_NO_COLOR}: ${_msg}"
}

# Displays an ERROR message
#
# $0: message to display
# $1: exit code
function _msg_error() {
  local _msg="${1}"
  local _rc=${2}

  echo -e "${_SCRIPT_NAME} | ${_RED}ERROR${_NO_COLOR}: ${_msg}"

  if [[ ${_rc} -ne 0 ]]; then
    exit "${_rc}"
  fi
}

# Prints the usage info of the script
function _usage() {
  cat <<_EOF_
NAME:
  ${_SCRIPT_NAME} - bootstrap an extension project

USAGE:
  ${_SCRIPT_NAME} [h|help]
  ${_SCRIPT_NAME} [g|gen|generate] /path/to/new/project

EXAMPLES:
  Display usage info:
    $ ${_SCRIPT_NAME} help

  Bootstrap a new project in the /path/foo/bar directory:
    $ ${_SCRIPT_NAME} generate /path/foo/bar
_EOF_
}

# Validates that expected env vars are set
function _validate_env_vars() {
  local _want_vars=(
    EXTENSION_NAME
    EXTENSION_TYPE
    EXTENSION_DESC
    ACTUATOR_NAME
    PRESERVE_HISTORY
  )

  for _var in "${_want_vars[@]}"; do
    if [ -z "${!_var}" ]; then
      _msg_error "Required var ${_var} is not set" 1
    fi
  done
}

# Bootstraps a new project
#
# $1: path in which to bootstrap project
function _bootstrap_project {
  local _dst_path="${1}"

  if [[ -z "${_dst_path}" ]]; then
    _msg_error "no destination path specified" 1
  fi

  local _extension_name_lc
  _extension_name_lc=$( echo "${EXTENSION_NAME}" | tr '[:upper:]-' '[:lower:]_' )

  _msg_info "Syncing project files to ${_dst_path} ..."
  rsync --archive \
        --delete-excluded \
        --exclude '*~' \
        --exclude '*.o' \
        --exclude 'bin/*' \
        --exclude bootstrap.sh \
        --exclude coverage.txt \
        --exclude 'generated.*.go' \
        --exclude 'hack/bootstrap.sh' \
        --exclude 'hack/bootstrap-vars.env' \
        --exclude 'images/*' \
        "${_PROJECT_DIR}/" "${_dst_path}"

  _msg_info "Fixing Go modules and packages ..."
  sed -i'' -e "s|gardener-extension-example|${EXTENSION_REPO}|g" "${_dst_path}/go.mod"
  sed -i'' -e "s|gardener-extension-example|${EXTENSION_REPO}|g" "${_dst_path}/internal/tools/go.mod"

  find "${_dst_path}" \
       -type f \
       -and -not -path "${_dst_path}/.git/*" \
       -iname "*.go" \
       -exec sed -i'' \
         -e "s|groupName=example.extensions.gardener.cloud|groupName=${EXTENSION_TYPE}.extensions.gardener.cloud|g" \
         -e "s|Usage:\s*:\"an example gardener extension\"|Usage: \"${EXTENSION_DESC}|g" \
         -e "s|Value:\s*\"gardener-extension-example\"|Value: \"${EXTENSION_NAME}\"|g" \
         -e "s|Name:\s*\"gardener-extension-example\"|Name: \"${EXTENSION_NAME}\"|g" \
         -e "s|Namespace = \"gardener_extension_example\"|Namespace = \"${_extension_name_lc}\"|g" \
         -e "s|Name = \"example\"|Name = \"${ACTUATOR_NAME}\"|g" \
         -e "s|Value:\s*\"gardener-extension-example-leader-election\"|Value: \"${EXTENSION_NAME}-leader-election\"|g" \
         -e "s|ExtensionType = \"example\"|ExtensionType = \"${EXTENSION_TYPE}\"|g" \
         -e "s|FinalizerSuffix = \"gardener-extension-example\"|FinalizerSuffix = \"${EXTENSION_NAME}\"|g" \
         -e "s|gardener-extension-example|${EXTENSION_REPO}|g" {} \;

  _msg_info "Fixing README ..."
  sed -i'' \
      -e "s|github.com/dnaeon/gardener-extension-example|${EXTENSION_REPO}|g" \
      -e "s|- type: example$|- type: ${EXTENSION_TYPE}|g" \
      -e "s|apiVersion: example.extensions.gardener.cloud/v1alpha1$|apiVersion: ${EXTENSION_TYPE}.extensions.gardener.cloud/v1alpha1|g" \
      -e "s|example   example   Succeeded   85m$|${EXTENSION_TYPE}   ${EXTENSION_TYPE}   Succeeded   85m|g" \
      -e "s|annotate extensions example gardener.cloud/operation=reconcile$|annotate extensions ${EXTENSION_TYPE} gardener.cloud/operation=reconcile|g" \
      -e "s|Extension/example   40s$|Extension/${EXTENSION_TYPE}   40s|g" \
      -e "s|repo provides the skeleton for an example Gardener extension.$|repo provides ${EXTENSION_DESC}.|g" \
      "${_dst_path}/README.md"

  sed -i'' '/BOOTSTRAP_DELETE_START/,/BOOTSTRAP_DELETE_END/d' \
      "${_dst_path}/README.md"

  _msg_info "Fixing Makefile targets ..."
    sed -i'' \
      -e "s|--provider-type example|--provider-type ${EXTENSION_TYPE}|g" \
      "${_dst_path}/Makefile"

  _msg_info "Fixing example manifests ..."
  env ext_type="${EXTENSION_TYPE}" api_version="${EXTENSION_TYPE}.extensions.gardener.cloud/v1alpha1" \
      go tool -modfile "${_TOOLS_MOD_FILE}" \
        yq -i '(.spec.extensions[0].type = env(ext_type)) | (.spec.extensions[0].providerConfig.apiVersion = env(api_version))' \
          "${_dst_path}/examples/shoot.yaml"

  env ext_type="${EXTENSION_TYPE}" \
      go tool -modfile "${_TOOLS_MOD_FILE}" \
        yq -i '(.spec.resources[0].type = env(ext_type))' \
          "${_dst_path}/examples/dev-setup/controllerregistration.yaml"

  env ext_type="${EXTENSION_TYPE}" \
      go tool -modfile "${_TOOLS_MOD_FILE}" \
        yq -i '(.spec.resources[0].type = env(ext_type))' \
          "${_dst_path}/examples/operator-extension/patches/extension.yaml"

  env ext_type="${EXTENSION_TYPE}" \
      go tool -modfile "${_TOOLS_MOD_FILE}" \
        yq -i '(.spec.resources[0].type = env(ext_type))' \
          "${_dst_path}/examples/operator-extension/base/extension.yaml"

  _msg_info "Fixing Dockerfile labels ..."
  sed -i'' \
      -e "s|org.opencontainers.image.source=\(.*\)|org.opencontainers.image.source=\"https://${EXTENSION_REPO}\"|g" \
      -e "s|org.opencontainers.image.title=\(.*\)|org.opencontainers.image.title=\"${EXTENSION_NAME}\"|g" \
      -e "s|org.opencontainers.image.description=\(.*\)|org.opencontainers.image.description=\"${EXTENSION_DESC}\"|g" \
      "${_dst_path}/Dockerfile"

  _msg_info "Fixing any remaining leftovers, hold tight ..."
  find "${_dst_path}" -type f \
       -and -not -path "${_dst_path}/.git/*" \
       -and -not -path "${_dst_path}/test/*" \
       -and -not -path "${_dst_path}/LICENSE" \
       -and -not -path "${_dst_path}/renovate.json" \
       -and -not -path "${_dst_path}/internal/tools/*" \
       -and -not -path "${_dst_path}/hack/*" \
       -and -not -path "${_dst_path}/CONTRIBUTING" \
       -and -not -path "${_dst_path}/VERSION" \
       -and -not -path "${_dst_path}/.golangci.yaml" \
       -and -not -path "${_dst_path}/go.sum" \
       -and -not -path "${_dst_path}/.github/ISSUE_TEMPLATE/*" \
       -exec \
         sed -i'' \
           -e "s|gardener-extension-example|${EXTENSION_NAME}|g" \
           -e "s|europe-docker.pkg.dev/gardener-project/public/gardener/extensions/example|europe-docker.pkg.dev/gardener-project/public/gardener/extensions/${EXTENSION_NAME}|g" \
           -e "s|dnaeon/gardener-extension-example|${EXTENSION_NAME}|g" \
           {} \;

  local _targets=(gotidy generate goimports-reviser check-helm check-examples)
  for _target in "${_targets[@]}"; do
    _msg_info "Running 'make ${_target}' ..."
    make -C "${_dst_path}" "${_target}"
  done

  _msg_info "All done."
  _msg_info "Project has been generated in ${_dst_path}"
  _msg_info "Next steps:"
  _msg_info " 1. Review the generated code"
  _msg_info " 2. Implement your extension logic in pkg/actuator/actuator.go"
  _msg_info " 3. Build the extension:"
  _msg_info "   $ make build"
  _msg_info " 4. Build a Docker image:"
  _msg_info "   $ make docker-build"
  _msg_info " 5. Deploy the extension in a local Gardener environment:"
  _msg_info "   $ make deploy"

  if [[ "${PRESERVE_HISTORY}" == "false" ]]; then
    _msg_info " 6. Initialize a new repo and commit"
    _msg_info "   $ git init -b main"
    _msg_info "   $ git add ."
    _msg_info "   $ git commit -m 'Initial commit'"
    rm -rf "${_dst_path}/.git"
  fi
}

# Main entrypoint
function _main() {
  if [[ $# -lt 1 ]]; then
    _usage
    exit 64  # EX_USAGE
  fi

  local _operation="${1}"
  case "${_operation}" in
    h|help)
      _usage
      exit 0
      ;;
    g|gen|generate)
      shift

      local _vars_file="${_SCRIPT_DIR}/bootstrap-vars.env"
      if [[ ! -f "${_vars_file}" ]]; then
        _msg_error "${_vars_file} not found" 1
      fi
      # shellcheck source=/dev/null
      source "${_vars_file}"
      _validate_env_vars

      _bootstrap_project "$@"
      ;;
    *)
      _msg_error "unknown command ${_operation}" 1
      exit 64  # EX_USAGE
      ;;
  esac
}

_main "$@"
