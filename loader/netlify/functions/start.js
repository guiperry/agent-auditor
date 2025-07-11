// netlify/functions/start.js
const AWS = require('aws-sdk');
const keys = require('../../config/keys');

// Helper function to start an EC2 instance
async function startEC2Instance(ec2, instanceId) {
  const params = {
    InstanceIds: [instanceId]
  };
  
  try {
    const result = await ec2.startInstances(params).promise();
    console.log(`Starting instance ${instanceId}`, result);
    return result;
  } catch (error) {
    console.error(`Error starting instance ${instanceId}:`, error);
    throw error;
  }
}

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
        error: 'Configuration error'
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
    await startEC2Instance(ec2, instanceId);
    
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
   - Go to Netlify dashboard → Site settings → Environment variables
   - Make sure there are no extra spaces or quotes in the values
   - Ensure the variable names are exactly NETLIFY_AWS_KEY_ID and NETLIFY_AWS_SECRET_KEY

2. Verify AWS credentials format:
   - Access Key ID should be 20 characters (e.g., AKIA...)
   - Secret Access Key should be 40 characters

3. Ensure the IAM user has EC2 permissions:
   - The user needs ec2:DescribeInstances, ec2:DescribeInstanceStatus, ec2:StartInstances permissions
   - Check the IAM console to verify the policy is attached

4. Region check:
   - Make sure your credentials are valid in the ${keys.region} region
   - The instance ID ${keys.instanceId} must exist in this region

5. After updating environment variables:
   - Go to Netlify dashboard → Deploys → Trigger deploy → Clear cache and deploy site
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
