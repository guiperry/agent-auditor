// netlify/functions/github-status.js
const axios = require('axios');
const ipStore = require('./ip-store');

// Function to check workflow status and get instance details
async function checkWorkflowStatus(token, owner, repo, runId) {
  try {
    // Get workflow run details
    const runUrl = `https://api.github.com/repos/${owner}/${repo}/actions/runs/${runId}`;
    
    const runResponse = await axios.get(
      runUrl,
      {
        headers: {
          'Accept': 'application/vnd.github.v3+json',
          'Authorization': `token ${token}`
        }
      }
    );
    
    const workflowRun = runResponse.data;
    
    // If the workflow is completed, get the instance IP from the artifacts or outputs
    if (workflowRun.status === 'completed' && workflowRun.conclusion === 'success') {
      // For simplicity, we'll check the instance status directly
      // In a real implementation, you might want to get this from workflow outputs or artifacts
      
      // Get instance status from a file in the repository or from EC2 directly
      // This is a simplified approach - in a real implementation, you'd need a more robust solution
      try {
        // Try to get the instance status from a status file or API
        const statusUrl = `https://raw.githubusercontent.com/${owner}/${repo}/${workflowRun.head_branch}/instance_status.json`;
        
        const statusResponse = await axios.get(statusUrl, {
          headers: {
            'Accept': 'application/json',
            'Cache-Control': 'no-cache'
          }
        });
        
        if (statusResponse.data && statusResponse.data.ip_address) {
          const ipAddress = statusResponse.data.ip_address;
          const redirectUrl = `http://${ipAddress}`;
          
          // Save the IP address to our storage
          console.log(`Saving IP address ${ipAddress} to storage`);
          await ipStore.saveIpAddress(ipAddress, true);
          
          return {
            success: true,
            status: 'completed',
            conclusion: 'success',
            isReady: true,
            publicIp: ipAddress,
            redirectUrl: redirectUrl,
            timestamp: statusResponse.data.timestamp
          };
        }
      } catch (statusError) {
        console.log('Could not get instance status from repository file:', statusError.message);
        // Continue with the workflow status only
      }
      
      // If we couldn't get the instance details, return the workflow status
      return {
        success: true,
        status: workflowRun.status,
        conclusion: workflowRun.conclusion,
        isReady: true,
        message: 'Workflow completed successfully, but instance details not available'
      };
    }
    
    // Return the workflow status
    return {
      success: true,
      status: workflowRun.status,
      conclusion: workflowRun.conclusion,
      isReady: workflowRun.status === 'completed' && workflowRun.conclusion === 'success',
      message: `Workflow ${workflowRun.status}, conclusion: ${workflowRun.conclusion || 'pending'}`
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

// Function to find the most recent workflow run
async function findRecentWorkflowRun(token, owner, repo, workflowId) {
  try {
    // Get recent workflow runs for the specific workflow
    const url = `https://api.github.com/repos/${owner}/${repo}/actions/workflows/${workflowId}/runs`;
    
    const response = await axios.get(
      url,
      {
        headers: {
          'Accept': 'application/vnd.github.v3+json',
          'Authorization': `token ${token}`
        }
      }
    );
    
    // Get the most recent run
    if (response.data.workflow_runs && response.data.workflow_runs.length > 0) {
      const mostRecentRun = response.data.workflow_runs[0];
      
      return {
        success: true,
        runId: mostRecentRun.id,
        status: mostRecentRun.status,
        conclusion: mostRecentRun.conclusion,
        createdAt: mostRecentRun.created_at
      };
    }
    
    return {
      success: false,
      message: 'No workflow runs found'
    };
  } catch (error) {
    console.error('Error finding recent workflow run:', error.message);
    
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
  
  // Check if GitHub token is available
  if (!githubToken) {
    return {
      statusCode: 500,
      headers: headers,
      body: JSON.stringify({
        status: 'error',
        message: 'GitHub token not configured',
        isReady: false
      })
    };
  }
  
  try {
    // Parse request body if it exists
    const requestBody = event.body ? JSON.parse(event.body) : {};
    const queryParams = event.queryStringParameters || {};
    
    // Get run ID from request body, query parameters, or find the most recent run
    let runId = requestBody.runId || queryParams.runId;
    
    if (!runId) {
      // Find the most recent workflow run
      const recentRun = await findRecentWorkflowRun(
        githubToken, 
        owner, 
        repo, 
        workflowId
      );
      
      if (!recentRun.success) {
        return {
          statusCode: 404,
          headers: headers,
          body: JSON.stringify({
            status: 'error',
            message: recentRun.message || 'No workflow runs found',
            isReady: false
          })
        };
      }
      
      runId = recentRun.runId;
    }
    
    // Check the workflow status
    const statusResult = await checkWorkflowStatus(
      githubToken, 
      owner, 
      repo, 
      runId
    );
    
    if (statusResult.success) {
      return {
        statusCode: 200,
        headers: headers,
        body: JSON.stringify({
          status: 'success',
          workflowStatus: statusResult.status,
          workflowConclusion: statusResult.conclusion,
          isReady: statusResult.isReady,
          publicIp: statusResult.publicIp,
          redirectUrl: statusResult.redirectUrl,
          state: statusResult.status === 'completed' ? 'running' : 'pending',
          message: statusResult.message || `Workflow status: ${statusResult.status}`
        })
      };
    } else {
      return {
        statusCode: statusResult.status || 500,
        headers: headers,
        body: JSON.stringify({
          status: 'error',
          message: `Failed to check workflow status: ${statusResult.error}`,
          error: statusResult.error,
          isReady: false
        })
      };
    }
  } catch (error) {
    console.error('Error in GitHub status function:', error);
    
    return {
      statusCode: 500,
      headers: headers,
      body: JSON.stringify({
        status: 'error',
        message: `Error: ${error.message}`,
        error: error.message,
        isReady: false
      })
    };
  }
};