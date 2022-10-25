const {abortFailedE2eCommand} = require("../constants");
const ci = require('../ci');
const fs = require('fs');

/**
 * Build additional info about failed e2e test
 * Contains information about
 *
 * @param {object} jobs - GitHub needsContext context
 * @returns {string}
 */
function buildFailedE2eTestAdditionalInfo({ needsContext, core, context }){
  core.debug("Start buildFailedE2eTestAdditionalInfo")
  const connectStrings = Object.getOwnPropertyNames(needsContext).
  filter((k) => k.startsWith('run_')).
  map((key, _i, _a) => {
    const result = needsContext[key].result;
    core.debug(`buildFailedE2eTestAdditionalInfo result for ${key}: result`)
    if (result === 'failure' || result === 'cancelled') {
      if (needsContext[key].outputs){
        const outputs = needsContext[key].outputs

        if(!outputs['failed_cluster_stayed']){
          return null;
        }

        // ci_commit_branch
        const connectStr = outputs['ssh_master_connection_string'] || ''
        const ranFor = outputs['ran_for'] || ''
        const runId = outputs['run_id'] || ''
        const issueNumber = outputs['issue_number'] || ''
        const artifactName = outputs['state_artifact_name'] || ''
        const clusterPrefix = needsContext[key].outputs['cluster_prefix'] || ''
        const ci_commit_ref_name = needsContext[key].outputs['ci_commit_ref_name'] || ''
        const pull_request_ref = needsContext[key].outputs['pull_request_ref'] || ''

        const argv = [
          abortFailedE2eCommand,
          ci_commit_ref_name,
          pull_request_ref,
          ranFor,
          runId,
          artifactName,
          clusterPrefix,
          issueNumber,
        ]

        core.debug(`result argv: ${JSON.stringify(argv)}`)

        const shouldArgc = argv.length
        const argc = argv.filter(v => !!v).length

        if (shouldArgc !== argc) {
          core.debug(`Incorrect outputs for ${key} ${shouldArgc} != ${argc}: ${JSON.stringify(argv)}; ${JSON.stringify(outputs)}`)
          core.error(`Incorrect outputs for ${key} ${shouldArgc} != ${argc}: ${JSON.stringify(argv)}; ${JSON.stringify(outputs)}`)
          return
        }

        const splitRunFor = ranFor.replace(';', ' ');
        const outConnectStr = connectStr ? `\`ssh -i ~/.ssh/e2e-id-rsa ${connectStr}\` - connect for debugging;` : '';

        return `
<!--- failed_clusters_start ${ranFor} -->
E2e for ${splitRunFor} was failed. Use:
  ${outConnectStr}

  \`${argv.join(' ')}\` - for abort failed cluster
<!--- failed_clusters_end ${ranFor} -->

`
      }
    }

    return null;
  }).filter((v) => !!v)

  if (connectStrings.length === 0) {
    core.debug("buildFailedE2eTestAdditionalInfo connection strings is empty")
    return "";
  }

  core.debug("buildFailedE2eTestAdditionalInfo was finished")
  return "\r\n" + connectStrings.join("\r\n") + "\r\n";
}

async function readConnectionScript({core, github, context}){
  core.debug(`SSH_CONNECT_STR_FILE ${process.env.SSH_CONNECT_STR_FILE}`);

  try {
    const data = fs.readFileSync(process.env.SSH_CONNECT_STR_FILE, 'utf8');
    core.setOutput('ssh_master_connection_string', data);
  } catch (err) {
    // this file can be not created
    core.warning(`Cannot read ssh connection file ${err.name}: ${err.message}`);
  }

  core.setOutput('failed_cluster_stayed', 'true');
}

module.exports = {
  buildFailedE2eTestAdditionalInfo,
  readConnectionScript,
}