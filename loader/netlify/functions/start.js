// netlify/functions/start.js
const AWS = require('aws-sdk');
const keys = require('../../config/keys');

exports.handler = async function(event, context) {
  // Log environment info for debugging (without exposing credentials)
  console.log(`Environment: ${process.env.NODE_ENV || 'not set'}`);
  console.log(`Netlify environment: ${process.env.NETLIFY ? 'true' : 'false'}`);
  console.log(`Context: ${process.env.CONTEXT || 'not set'}`);
  console.log(`AWS Region: ${keys.region}`);
  console.log(`Instance ID: ${keys.instanceId}`);
  console.log(`Access Key ID provided: ${keys.accessKeyId ? 'Yes (masked)' : 'No'}`);
  console.log(`Secret Access Key provided: ${keys.secretAccessKey ? 'Yes (masked)' : 'No'}`);

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
        error: 'Configuration error'
      })
    };
  }

  // Configure AWS SDK with our custom keys configuration
  AWS.config.update({
    region: keys.region,
    credentials: new AWS.Credentials({
      accessKeyId: keys.accessKeyId,
      secretAccessKey: keys.secretAccessKey
    })
  });

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
          error: 'Configuration error'
        })
      };
    }

    console.log(`🤖 Robot game initiated! Starting EC2 instance: ${instanceId}`);

    // Check if the instance is already running
    const statusResponse = await ec2.describeInstanceStatus({
      InstanceIds: [instanceId],
      IncludeAllInstances: true
    }).promise();

    const instanceInfo = statusResponse.InstanceStatuses[0];
    
    // If instance exists and is already running, return its details
    if (instanceInfo && 
        (instanceInfo.InstanceState.Name === 'running' || 
         instanceInfo.InstanceState.Name === 'pending')) {
      
      // Get public IP address
      const describeResponse = await ec2.describeInstances({
        InstanceIds: [instanceId]
      }).promise();
      
      const instance = describeResponse.Reservations[0].Instances[0];
      const publicIp = instance.PublicIpAddress;
      const redirectUrl = publicIp ? `http://${publicIp}:8080` : null;
      
      console.log(`✅ Instance ${instanceId} is already ${instanceInfo.InstanceState.Name}`);
      
      return {
        statusCode: 200,
        headers: headers,
        body: JSON.stringify({
          status: 'already_running',
          instanceId: instanceId,
          publicIp: publicIp,
          redirectUrl: redirectUrl,
          isReady: instanceInfo.InstanceState.Name === 'running',
          state: instanceInfo.InstanceState.Name,
          message: `Instance is already ${instanceInfo.InstanceState.Name}`
        })
      };
    }
    
    // Start the instance if it's not running
    await ec2.startInstances({
      InstanceIds: [instanceId]
    }).promise();
    
    console.log(`✅ Instance ${instanceId} start command sent`);
    
    return {
      statusCode: 200,
      headers: headers,
      body: JSON.stringify({
        status: 'starting',
        instanceId: instanceId,
        isReady: false,
        state: 'pending',
        message: 'Instance is starting',
        estimatedTime: '2-3 minutes'
      })
    };
  } catch (error) {
    console.error('Error starting instance:', error);
    
    // Provide more specific guidance for common errors
    let errorMessage = `Error starting instance: ${error.message}`;
    let troubleshootingSteps = '';
    
    if (error.code === 'AuthFailure' || error.message.includes('validate the provided access credentials')) {
      troubleshootingSteps = `
Please check the following:
1. Verify that NETLIFY_AWS_KEY_ID and NETLIFY_AWS_SECRET_KEY are correctly set in Netlify environment variables
2. Ensure the IAM user has EC2 permissions (ec2:DescribeInstances, ec2:StartInstances)
3. Check if the access keys are still active and not expired
4. After updating environment variables, redeploy your Netlify site
`;
    } else if (error.code === 'InvalidInstanceID.NotFound') {
      troubleshootingSteps = `
Please check the following:
1. Verify that NETLIFY_EC2_INSTANCE_ID is correctly set in Netlify environment variables
2. Ensure the instance exists in the specified AWS region (${keys.region})
3. Check if the instance has been terminated or deleted
`;
    } else if (error.code === 'UnauthorizedOperation') {
      troubleshootingSteps = `
Please check the following:
1. The IAM user lacks permission to perform this action
2. Add the required EC2 permissions to your IAM user
`;
    }
    
    return {
      statusCode: 500,
      headers: headers,
      body: JSON.stringify({
        status: 'error',
        message: errorMessage,
        troubleshooting: troubleshootingSteps.trim(),
        error: error.toString()
      })
    };
  }
};
