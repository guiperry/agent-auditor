// netlify/functions/status.js
const AWS = require('aws-sdk');
const keys = require('../../config/keys');

exports.handler = async function(event, context) {
  // Log environment info for debugging (without exposing full credentials)
  console.log(`Environment: ${process.env.NODE_ENV || 'not set'}`);
  console.log(`Netlify environment: ${process.env.NETLIFY ? 'true' : 'false'}`);
  console.log(`Context: ${process.env.CONTEXT || 'not set'}`);
  console.log(`AWS Region: ${keys.region}`);
  console.log(`Instance ID: ${keys.instanceId}`);
  
  // Log partial credentials for debugging (first 4 chars only)
  const accessKeyPrefix = keys.accessKeyId ? keys.accessKeyId.substring(0, 4) + '...' : 'Not provided';
  const secretKeyPrefix = keys.secretAccessKey ? keys.secretAccessKey.substring(0, 4) + '...' : 'Not provided';
  console.log(`Access Key ID: ${accessKeyPrefix}`);
  console.log(`Secret Access Key: ${secretKeyPrefix}`);
  
  // Log environment variable names that are available (without values)
  console.log('Available environment variables (names only):');
  Object.keys(process.env)
    .filter(key => key.includes('AWS') || key.includes('NETLIFY') || key === 'NODE_ENV' || key === 'CONTEXT')
    .forEach(key => console.log(`  - ${key}`));

  // Validate credentials before proceeding
  if (!keys.accessKeyId || !keys.secretAccessKey) {
    return {
      statusCode: 500,
      headers: {
        'Access-Control-Allow-Origin': '*',
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        status: 'error',
        message: 'AWS credentials are missing. Please check your environment variables.',
        error: 'Configuration error',
        isReady: false,
        state: 'error'
      })
    };
  }

  // Configure AWS SDK with direct credential assignment
  // This is the simplest and most reliable approach
  AWS.config.update({
    region: keys.region,
    accessKeyId: keys.accessKeyId,
    secretAccessKey: keys.secretAccessKey
  });
  
  // Log AWS SDK version for debugging
  console.log(`AWS SDK Version: ${AWS.VERSION}`);
  
  // Create EC2 service object
  const ec2 = new AWS.EC2();
  const instanceId = keys.instanceId;

  // CORS headers for browser requests
  const headers = {
    'Access-Control-Allow-Origin': '*', // Will be configured in Netlify
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

  try {
    // Validate configuration
    if (!keys.region || !instanceId) {
      return {
        statusCode: 500,
        headers: headers,
        body: JSON.stringify({
          status: 'error',
          message: 'Missing required configuration (region, instanceId)',
          error: 'Configuration error',
          isReady: false,
          state: 'error'
        })
      };
    }

    // Get instance status
    const statusResponse = await ec2.describeInstanceStatus({
      InstanceIds: [instanceId],
      IncludeAllInstances: true
    }).promise();

    if (!statusResponse.InstanceStatuses || statusResponse.InstanceStatuses.length === 0) {
      return {
        statusCode: 404,
        headers: headers,
        body: JSON.stringify({
          status: 'error',
          message: 'Instance not found',
          isReady: false,
          state: 'unknown'
        })
      };
    }

    const instanceInfo = statusResponse.InstanceStatuses[0];
    const instanceState = instanceInfo.InstanceState.Name;
    
    // Check if instance is running and status checks have passed
    const isRunning = instanceState === 'running';
    const systemStatus = instanceInfo.SystemStatus ? instanceInfo.SystemStatus.Status : 'not_available';
    const instanceStatus = instanceInfo.InstanceStatus ? instanceInfo.InstanceStatus.Status : 'not_available';
    const allChecksOk = isRunning && systemStatus === 'ok' && instanceStatus === 'ok';
    
    // If instance is running, get its public IP
    let publicIp = null;
    let redirectUrl = null;
    
    if (isRunning) {
      const describeResponse = await ec2.describeInstances({
        InstanceIds: [instanceId]
      }).promise();
      
      const instance = describeResponse.Reservations[0].Instances[0];
      publicIp = instance.PublicIpAddress;
      
      if (publicIp) {
        redirectUrl = `http://${publicIp}:8080`;
        console.log(`🎉 Instance ${instanceId} is ready! Public IP: ${publicIp}`);
      } else {
        // Fallback IP address if instance IP is not available
        const FALLBACK_IP = "3.146.37.27:8080";
        redirectUrl = `http://${FALLBACK_IP}`;
        console.log(`🎉 Instance ${instanceId} is ready! Using fallback IP: ${FALLBACK_IP}`);
      }
    } else {
      console.log(`⏳ Instance ${instanceId} status: ${instanceState}`);
    }
    
    return {
      statusCode: 200,
      headers: headers,
      body: JSON.stringify({
        status: 'success',
        instanceId: instanceId,
        state: instanceState,
        systemStatus: systemStatus,
        instanceStatus: instanceStatus,
        isReady: allChecksOk,
        publicIp: publicIp,
        redirectUrl: redirectUrl,
        message: allChecksOk ? 'Instance is ready' : `Instance is ${instanceState}`
      })
    };
  } catch (error) {
    console.error('Error checking instance status:', error);
    
    return {
      statusCode: 500,
      headers: headers,
      body: JSON.stringify({
        status: 'error',
        message: `Error checking instance status: ${error.message}`,
        error: error.toString(),
        isReady: false,
        state: 'error'
      })
    };
  }
};
