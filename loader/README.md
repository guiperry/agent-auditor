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

- `NETLIFY_AWS_REGION` - AWS region where your EC2 instance is located (e.g., `us-east-1`)
- `NETLIFY_AWS_KEY_ID` - Your AWS access key ID
- `NETLIFY_AWS_SECRET_KEY` - Your AWS secret access key
- `NETLIFY_EC2_INSTANCE_ID` - ID of the EC2 instance to manage (e.g., `i-0123456789abcdef0`)

> **Important:** We use custom environment variable names to avoid Netlify's reserved environment variable restrictions. Netlify reserves standard AWS environment variable names (`AWS_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`) for its own use.

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