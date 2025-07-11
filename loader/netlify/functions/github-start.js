// netlify/functions/github-start.js
const axios = require('axios');
const { v4: uuidv4 } = require('uuid');

// Function to trigger GitHub workflow
async function triggerGitHubWorkflow(token, owner, repo, workflow_id, ref = 'main') {
  try {
    const url = `https://api.github.com/repos/${owner}/${repo}/actions/workflows/${workflow_id}/dispatches`;
    
    // Generate a unique caller ID for tracking
    const callerId = uuidv4();
    
    const response = await axios.post(
      url,
      {
        ref: ref,
        inputs: {
          caller_id: callerId
        }
      },
      {
        headers: {
          'Accept': 'application/vnd.github.v3+json',
          'Authorization': `token ${token}`,
          'Content-Type': 'application/json'
        }
      }
    );
    
    return {
      success: true,
      status: response.status,
      callerId: callerId
    };
  } catch (error) {
    console.error('Error triggering GitHub workflow:', error.message);
    
    return {
      success: false,
      error: error.message,
      status: error.response ? error.response.status : 500
    };
  }
}

// Function to check workflow status
async function checkWorkflowStatus(token, owner, repo, callerId) {
  try {
    // Get recent workflow runs
    const url = `https://api.github.com/repos/${owner}/${repo}/actions/runs`;
    
    const response = await axios.get(
      url,
      {
        headers: {
          'Accept': 'application/vnd.github.v3+json',
          'Authorization': `token ${token}`
        }
      }
    );
    
    // Find the workflow run with our caller ID
    // Note: This is a simplification. In a real implementation, you'd need to
    // check the workflow run's inputs to match the caller_id
    const workflowRun = response.data.workflow_runs.find(run => 
      run.name === 'Start EC2 Instance' && 
      new Date(run.created_at) > new Date(Date.now() - 10 * 60 * 1000) // Within last 10 minutes
    );
    
    if (!workflowRun) {
      return {
        success: false,
        status: 'not_found',
        message: 'Workflow run not found'
      };
    }
    
    return {
      success: true,
      status: workflowRun.status,
      conclusion: workflowRun.conclusion,
      id: workflowRun.id,
      url: workflowRun.html_url
    };
  } catch (error) {
    console.error('Error checking workflow status:', error.message);
    
    return {
      success: false,
      error: error.message,
      status: error.response ? error.response.status : 500
    };
  }
}

exports.handler = async function(event, context) {
  // CORS headers for browser requests
  const headers = {
    'Access-Control-Allow-Origin': '*',
    'Access-Control-Allow-Headers': 'Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token',
    'Access-Control-Allow-Methods': 'GET,POST,OPTIONS',
    'Content-Type': 'application/json'
  };
  
  // Handle preflight OPTIONS requests
  if (event.httpMethod === 'OPTIONS') {
    return {
      statusCode: 200,
      headers: headers,
      body: JSON.stringify({ message: 'CORS preflight' })
    };
  }
  
  // Get GitHub token from environment variables
  const githubToken = process.env.GITHUB_TOKEN;
  const owner = process.env.GITHUB_OWNER || 'your-github-username';
  const repo = process.env.GITHUB_REPO || 'agent-auditor';
  const workflowId = process.env.GITHUB_WORKFLOW_ID || 'start-ec2-instance.yml';
  const ref = process.env.GITHUB_REF || 'main';
  
  // Check if GitHub token is available
  if (!githubToken) {
    return {
      statusCode: 500,
      headers: headers,
      body: JSON.stringify({
        status: 'error',
        message: 'GitHub token not configured',
        troubleshooting: `
Please add the following environment variables in your Netlify dashboard:
1. GITHUB_TOKEN - A personal access token with 'workflow' scope
2. GITHUB_OWNER - Your GitHub username or organization name
3. GITHUB_REPO - The repository name (default: agent-auditor)
4. GITHUB_WORKFLOW_ID - The workflow file name (default: start-ec2-instance.yml)
5. GITHUB_REF - The branch or tag to use (default: main)
`
      })
    };
  }
  
  try {
    // Parse request body if it exists
    const requestBody = event.body ? JSON.parse(event.body) : {};
    
    // Check if this is a status check request
    if (requestBody.action === 'check_status' && requestBody.callerId) {
      const statusResult = await checkWorkflowStatus(
        githubToken, 
        owner, 
        repo, 
        requestBody.callerId
      );
      
      return {
        statusCode: statusResult.success ? 200 : 500,
        headers: headers,
        body: JSON.stringify({
          status: statusResult.success ? 'success' : 'error',
          workflowStatus: statusResult.status,
          workflowConclusion: statusResult.conclusion,
          message: statusResult.message || `Workflow status: ${statusResult.status}`,
          workflowId: statusResult.id,
          workflowUrl: statusResult.url
        })
      };
    }
    
    // Trigger the GitHub workflow
    console.log(`🚀 Triggering GitHub workflow to start EC2 instance`);
    const result = await triggerGitHubWorkflow(
      githubToken, 
      owner, 
      repo, 
      workflowId,
      ref
    );
    
    if (result.success) {
      return {
        statusCode: 200,
        headers: headers,
        body: JSON.stringify({
          status: 'starting',
          message: 'GitHub workflow triggered successfully',
          callerId: result.callerId,
          isReady: false,
          state: 'pending',
          estimatedTime: '2-3 minutes'
        })
      };
    } else {
      return {
        statusCode: result.status || 500,
        headers: headers,
        body: JSON.stringify({
          status: 'error',
          message: `Failed to trigger GitHub workflow: ${result.error}`,
          error: result.error,
          troubleshooting: `
Please check the following:
1. Verify that GITHUB_TOKEN has the 'workflow' scope
2. Ensure the repository and workflow file exist
3. Check that the token has permission to trigger workflows
4. Verify the branch/ref exists in the repository
`
        })
      };
    }
  } catch (error) {
    console.error('Error in GitHub workflow function:', error);
    
    return {
      statusCode: 500,
      headers: headers,
      body: JSON.stringify({
        status: 'error',
        message: `Error: ${error.message}`,
        error: error.message
      })
    };
  }
};