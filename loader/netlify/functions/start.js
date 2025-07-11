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

// Function to verify AWS credentials and try fallback options
async function verifyCredentials(credentials) {
  console.log('Verifying AWS credentials...');
  
  try {
    // Try to use the credentials to make a simple AWS call
    const sts = new AWS.STS({
      region: credentials.region,
      accessKeyId: credentials.accessKeyId,
      secretAccessKey: credentials.secretAccessKey
    });
    
    // GetCallerIdentity is a simple call that will validate credentials
    const identity = await sts.getCallerIdentity().promise();
    console.log(`✅ Credentials verified. User: ${identity.Arn}`);
    return { valid: true, identity };
  } catch (error) {
    console.error('❌ Credential verification failed:', error.message);
    return { valid: false, error };
  }
}

// Function to use a fallback strategy for EC2 access
async function getEC2WithFallbackStrategy(credentials) {
  // First, try with the provided credentials
  try {
    // Configure AWS SDK with direct credential assignment
    AWS.config.update({
      region: credentials.region,
      accessKeyId: credentials.accessKeyId,
      secretAccessKey: credentials.secretAccessKey
    });
    
    // Verify the credentials
    const verificationResult = await verifyCredentials(credentials);
    
    if (verificationResult.valid) {
      // Credentials are valid, create and return EC2 client
      return { 
        success: true, 
        ec2: new AWS.EC2(),
        message: 'Using provided credentials'
      };
    }
  } catch (error) {
    console.error('Error using provided credentials:', error);
  }
  
  console.log('Trying fallback strategies...');
  
  // Fallback 1: Try using the AWS SDK's default credential provider chain
  try {
    console.log('Fallback 1: Using AWS SDK default credential provider chain');
    AWS.config.update({ region: credentials.region });
    
    // Clear any previously set credentials
    AWS.config.credentials = null;
    
    // Create a new EC2 client without explicit credentials
    const ec2 = new AWS.EC2();
    
    // Test if it works by making a simple call
    await ec2.describeRegions().promise();
    
    return { 
      success: true, 
      ec2,
      message: 'Using AWS SDK default credential provider chain'
    };
  } catch (error) {
    console.error('Fallback 1 failed:', error.message);
  }
  
  // Fallback 2: Try using environment credentials directly
  try {
    console.log('Fallback 2: Using standard AWS environment variables directly');
    
    const envCredentials = {
      region: process.env.AWS_REGION || process.env.AWS_DEFAULT_REGION || credentials.region,
      accessKeyId: process.env.AWS_ACCESS_KEY_ID,
      secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY
    };
    
    if (envCredentials.accessKeyId && envCredentials.secretAccessKey) {
      AWS.config.update(envCredentials);
      
      const ec2 = new AWS.EC2();
      await ec2.describeRegions().promise();
      
      return { 
        success: true, 
        ec2,
        message: 'Using standard AWS environment variables'
      };
    } else {
      console.log('Standard AWS environment variables not found');
    }
  } catch (error) {
    console.error('Fallback 2 failed:', error.message);
  }
  
  // Fallback 3: Use a mock EC2 client for development/testing
  if (process.env.NODE_ENV === 'development' || process.env.MOCK_AWS === 'true') {
    console.log('Fallback 3: Using mock EC2 client for development');
    
    // Create a mock EC2 client that simulates responses
    const mockEC2 = {
      startInstances: () => ({
        promise: () => Promise.resolve({
          StartingInstances: [{ InstanceId: credentials.instanceId, CurrentState: { Name: 'pending' } }]
        })
      }),
      describeInstanceStatus: () => ({
        promise: () => Promise.resolve({
          InstanceStatuses: [{
            InstanceId: credentials.instanceId,
            InstanceState: { Name: 'running' },
            SystemStatus: { Status: 'ok' },
            InstanceStatus: { Status: 'ok' }
          }]
        })
      }),
      describeInstances: () => ({
        promise: () => Promise.resolve({
          Reservations: [{
            Instances: [{
              InstanceId: credentials.instanceId,
              PublicIpAddress: 'agent-auditor.fly.dev' // Redirect to the actual application
            }]
          }]
        })
      }),
      describeRegions: () => ({
        promise: () => Promise.resolve({ Regions: [{ RegionName: credentials.region }] })
      })
    };
    
    return { 
      success: true, 
      ec2: mockEC2,
      message: 'Using mock EC2 client for development',
      isMock: true
    };
  }
  
  // All fallbacks failed
  return { 
    success: false, 
    message: 'All credential strategies failed',
    error: 'Unable to establish valid AWS credentials'
  };
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

  // CORS headers for browser requests
  const headers = {
    'Access-Control-Allow-Origin': '*',
    'Access-Control-Allow-Headers': 'Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token',
    'Access-Control-Allow-Methods': 'GET,POST,OPTIONS',
    'Content-Type': 'application/json'
  };

  // Log AWS SDK version for debugging
  console.log(`AWS SDK Version: ${AWS.VERSION}`);
  
  // Prepare credentials object
  const credentials = {
    region: keys.region,
    accessKeyId: keys.accessKeyId,
    secretAccessKey: keys.secretAccessKey,
    instanceId: keys.instanceId
  };
  
  // Try to get a working EC2 client using our fallback strategy
  const ec2Result = await getEC2WithFallbackStrategy(credentials);
  
  if (!ec2Result.success) {
    return {
      statusCode: 500,
      headers,
      body: JSON.stringify({
        status: 'error',
        message: 'AWS credentials could not be validated after trying multiple strategies.',
        details: ec2Result.message,
        error: ec2Result.error,
        troubleshooting: `
Please check the following:
1. Verify that NETLIFY_AWS_KEY_ID and NETLIFY_AWS_SECRET_KEY are correctly set in Netlify environment variables
2. Ensure the IAM user has EC2 permissions
3. Check if the access keys are still active and not expired
4. Try setting standard AWS environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY) as a fallback
5. After updating environment variables, redeploy your Netlify site

For testing without AWS credentials:
- Set MOCK_AWS=true in your Netlify environment variables to use a mock AWS client
- This will allow you to test the UI flow without real AWS credentials
`
      })
    };
  }
  
  console.log(`✅ EC2 client initialized: ${ec2Result.message}`);
  
  // Get the EC2 client and instance ID
  const ec2 = ec2Result.ec2;
  const instanceId = keys.instanceId;
  
  // If we're using a mock client, log that information
  if (ec2Result.isMock) {
    console.log('⚠️ Using mock EC2 client - responses will be simulated');
  }

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
      // Use HTTPS for agent-auditor.fly.dev, otherwise use HTTP with port 80
      const redirectUrl = publicIp ? 
        (publicIp === 'agent-auditor.fly.dev' ? `https://${publicIp}` : `http://${publicIp}:80`) : 
        null;
      
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
