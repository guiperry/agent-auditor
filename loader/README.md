# 🤖 Build-a-Bot: Interactive Loading Game for Agent Auditor

An engaging robot-building game that entertains users while Netlify Functions perform backend operations. This project uses Netlify Functions to manage EC2 instances, providing a simplified serverless architecture.

## 📋 Overview

Build-a-Bot provides an interactive loading experience that keeps users engaged during the wait time for AWS resources to initialize. Instead of showing a boring loading spinner, users can build their own robot while Netlify Functions start an EC2 instance in the background.

## 🏗️ Netlify Functions Architecture

This application uses Netlify Functions with these endpoints:

| Method | Path                        | Function    | Purpose                    |
|--------|-----------------------------|-------------|----------------------------|
| POST   | /.netlify/functions/start   | start.js    | Start EC2 instance         |
| GET    | /.netlify/functions/status  | status.js   | Check instance status      |

## 🚀 Deployment Steps

### Step 1: Install Dependencies

1. Navigate to the loader directory:
   ```bash
   cd loader
   npm install
   ```

### Step 2: Configure Environment Variables

Set these environment variables in your Netlify dashboard (Site settings > Environment variables):

- `NETLIFY_AWS_REGION` - AWS region where your EC2 instance is located (e.g., `us-east-2`)
- `NETLIFY_AWS_KEY_ID` - Your AWS access key ID
- `NETLIFY_AWS_SECRET_KEY` - Your AWS secret access key
- `NETLIFY_EC2_INSTANCE_ID` - ID of the EC2 instance to manage (e.g., `i-0123456789abcdef0`)

> **Important:** We use custom environment variable names to avoid Netlify's reserved environment variable restrictions. Netlify reserves standard AWS environment variable names (`AWS_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`) for its own use.

#### AWS SDK Configuration

The application uses the AWS SDK for JavaScript to interact with EC2 instances. The SDK is configured with a robust fallback strategy to handle credential issues:

```javascript
// Try to get a working EC2 client using our fallback strategy
const ec2Result = await getEC2WithFallbackStrategy(credentials);

// Get the EC2 client if successful
const ec2 = ec2Result.ec2;
```

##### Credential Fallback Strategy

The application implements a multi-tier fallback strategy for AWS credentials:

1. **Primary Strategy**: Use the Netlify environment variables
   ```javascript
   AWS.config.update({
     region: keys.region,
     accessKeyId: keys.accessKeyId,
     secretAccessKey: keys.secretAccessKey
   });
   ```

2. **Fallback 1**: Use AWS SDK's default credential provider chain
   - This tries environment variables, shared credentials file, and instance profiles

3. **Fallback 2**: Try standard AWS environment variables directly
   ```javascript
   const envCredentials = {
     region: process.env.AWS_REGION || process.env.AWS_DEFAULT_REGION,
     accessKeyId: process.env.AWS_ACCESS_KEY_ID,
     secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY
   };
   ```

4. **Fallback 3**: Use a mock EC2 client for development/testing
   - This provides simulated responses for local development

This multi-tier approach ensures maximum reliability when dealing with AWS credentials in a serverless environment.

#### Troubleshooting AWS Credential Issues

If you see the error: `"AWS was not able to validate the provided access credentials"`, follow these steps:

1. **Verify Environment Variables in Netlify**:
   - Go to Netlify dashboard → Site settings → Environment variables
   - Ensure all four variables are set correctly:
     - `NETLIFY_AWS_REGION` (e.g., `us-east-2`)
     - `NETLIFY_AWS_KEY_ID` (should be 20 characters, starting with `AKIA`)
     - `NETLIFY_AWS_SECRET_KEY` (should be 40 characters)
     - `NETLIFY_EC2_INSTANCE_ID` (e.g., `i-0123456789abcdef0`)
   - **Important**: Make sure there are no extra spaces, quotes, or newlines in the values
   - If you copy-pasted the values, try typing them manually instead

2. **Try Standard AWS Environment Variables**:
   - If the Netlify-specific variables aren't working, try setting these standard variables:
     - `AWS_REGION`
     - `AWS_ACCESS_KEY_ID`
     - `AWS_SECRET_ACCESS_KEY`
     - `EC2_INSTANCE_ID`
   - Our code will fall back to these if the Netlify-specific ones aren't found

3. **Verify AWS IAM Permissions**:
   - The IAM user associated with your credentials needs these permissions:
     ```json
     {
       "Version": "2012-10-17",
       "Statement": [
         {
           "Effect": "Allow",
           "Action": [
             "ec2:DescribeInstances",
             "ec2:DescribeInstanceStatus",
             "ec2:StartInstances",
             "ec2:StopInstances"
           ],
           "Resource": "*"
         }
       ]
     }
     ```
   - You can also attach the `AmazonEC2FullAccess` managed policy for testing

4. **Check AWS Credential Status**:
   - Verify the credentials are active and not expired
   - Ensure the IAM user has not been deleted or disabled
   - Try creating new access keys in the AWS IAM console
   - Make sure the keys are activated (sometimes new keys need activation)

5. **Region and Instance Check**:
   - Confirm your EC2 instance exists in the specified region
   - Verify the instance ID is correct and the instance is not terminated
   - Try accessing the instance through the AWS console to confirm it exists

6. **Redeploy After Changes**:
   - After updating environment variables, trigger a new deployment:
     - Go to Netlify dashboard → Deploys → Trigger deploy → Clear cache and deploy site
   - This ensures your changes take effect

7. **Check Netlify Logs**:
   - Go to Netlify dashboard → Functions → Select your function → View logs
   - Look for any error messages or warnings related to AWS credentials

8. **Test Locally First**:
   - Set the environment variables locally and test with `netlify dev`
   - This can help isolate whether the issue is with Netlify or your code

9. **Enable Development Mode for Testing**:
   - Set the `CONTEXT` environment variable to `dev` in Netlify:
     ```
     CONTEXT=dev
     ```
   - This will activate the mock EC2 client that simulates AWS responses
   - You can test your application without real AWS credentials
   - **Note**: This is only for testing the UI flow, not for production use

10. **Last Resort - Hardcode for Testing**:
   - As a temporary measure for testing only, you can hardcode credentials:
     ```javascript
     // TEMPORARY FOR TESTING ONLY - REMOVE AFTER TESTING
     if (!accessKeyId) accessKeyId = 'YOUR_ACCESS_KEY_ID';
     if (!secretAccessKey) secretAccessKey = 'YOUR_SECRET_ACCESS_KEY';
     ```
   - **IMPORTANT**: Remove hardcoded credentials immediately after testing!

### Step 3: Deploy to Netlify

1. Connect your GitHub repository to Netlify
2. Set the base directory to `loader`
3. The build settings are automatically configured via `netlify.toml`
4. Deploy the site

### Step 4: Local Development (Optional)

For local testing:
```bash
cd loader
npm run dev
```

### Step 5: Test the Flow

1. User visits your hosted Netlify page
2. Page automatically calls `/.netlify/functions/start` endpoint
3. Netlify Function starts EC2 instance using AWS SDK
4. Page polls `/.netlify/functions/status` endpoint every 5 seconds
5. User plays with the robot-building game while waiting
6. When ready, the page shows "Go to Instance" button
7. User can click to redirect to their EC2 instance

## 🔒 Security Enhancements

For production environments, consider:

1. **Restrict CORS origins** in `netlify.toml`:
   ```toml
   [[headers]]
     for = "/.netlify/functions/*"
     [headers.values]
       Access-Control-Allow-Origin = "https://your-app.netlify.app"
   ```

2. **Add configuration validation** in your functions:
   ```javascript
   if (!keys.region || !keys.instanceId) {
     return {
       statusCode: 500,
       body: JSON.stringify({ error: 'Missing required configuration' })
     };
   }
   ```

3. **Use IAM roles with minimal permissions** for your AWS credentials

## ✨ Features

This Netlify Functions setup provides:
- Real-time status monitoring via serverless functions
- Smooth user experience with no API Gateway complexity
- Automatic redirection when EC2 instance is ready
- Comprehensive error handling
- Built-in CORS support
- Simplified deployment process
- Local development support with `netlify dev`

## 🚀 Auto-Redirect Flow

The game will automatically redirect to your EC2 instance as soon as it's ready:

1. **Instance Ready Detection**: Lambda function reports when the instance is running
2. **5-Second Countdown**: Shows a countdown timer with "Redirecting in X seconds..."
3. **User Options During Countdown**:
   - "Go Now" button - Immediate redirect
   - "Cancel" button - Stops auto-redirect
   - "Copy IP" button - Copies IP to clipboard
4. **Automatic Redirect**: After countdown, opens EC2 instance in a new tab
5. **Visual Feedback**: Shows a spinning robot overlay during redirect

## 🎛️ User Control Features

- **Countdown Timer**: 5-second warning before redirect
- **Cancel Option**: Users can stop auto-redirect if they want to stay on the page
- **Manual Redirect**: "Go Now" button for immediate redirect
- **Visual Loading**: Spinning robot animation during redirect
- **New Tab Opening**: Doesn't navigate away from the game page

## 🔧 Customization Options

### For immediate redirect (no countdown):
```javascript
// Replace showRedirectCountdown with:
function showRedirectCountdown(data) {
    updateStatus('✅ EC2 Instance Ready! Redirecting now... 🚀');
    redirectToInstance(data.redirectUrl);
}
```

### For a longer countdown:
```javascript
let countdown = 10; // 10 seconds instead of 5
```

## 🛡️ Error Handling

The system handles cases where:
- No redirect URL is available
- User cancels the redirect
- Network errors during status checking

This creates a smooth, automatic experience while still giving users control over the redirect behavior. The game keeps them engaged during the wait time, then seamlessly transitions them to their EC2 instance!

## 🏗️ Architecture Benefits

The Netlify Functions implementation provides several advantages over the previous AWS Lambda + API Gateway setup:

### Simplified Infrastructure
- **Single Platform**: Everything hosted on Netlify (static files + serverless functions)
- **No API Gateway**: Direct function endpoints eliminate additional configuration
- **Unified Deployment**: One deployment process for frontend and backend

### Developer Experience
- **Local Development**: Use `netlify dev` to test functions locally
- **Environment Variables**: Managed through Netlify dashboard
- **Automatic HTTPS**: Built-in SSL certificates
- **Git Integration**: Automatic deployments from repository

### Cost Efficiency
- **Reduced Services**: Fewer AWS services to manage and pay for
- **Netlify Free Tier**: Generous limits for small to medium applications
- **Pay-per-Use**: Functions only run when needed

### Maintenance
- **Simplified Monitoring**: Single platform for logs and analytics
- **Easier Updates**: Direct file updates without Lambda deployment packages
- **Version Control**: Functions are part of your repository