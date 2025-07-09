# 🤖 Build-a-Bot: Interactive Loading Game for AWS Lambda

An engaging robot-building game that entertains users while AWS Lambda functions perform backend operations. This project uses Amazon API Gateway to connect the frontend game with Lambda functions that manage EC2 instances.

## 📋 Overview

Build-a-Bot provides an interactive loading experience that keeps users engaged during the wait time for AWS resources to initialize. Instead of showing a boring loading spinner, users can build their own robot while the Lambda function starts an EC2 instance in the background.

## 🏗️ AWS API Gateway Setup

You'll need to configure Amazon API Gateway with these endpoints:

| Method | Path     | Integration     | CORS  |
|--------|----------|----------------|-------|
| POST   | /start   | Lambda Function | Yes   |
| GET    | /status  | Lambda Function | Yes   |

## 🚀 Deployment Steps

### Step 1: Deploy Lambda Function

1. Update your Lambda function with the code from `EC2_Lambda.py`
2. Set environment variables:
   - `REGION` - AWS region where your EC2 instance is located
   - `INSTANCE_ID` - ID of the EC2 instance to manage
3. Ensure Lambda has proper IAM permissions to manage EC2 instances

### Step 2: Create API Gateway

1. Go to AWS API Gateway console
2. Create a new REST API
3. Create resources `/start` and `/status`
4. Add POST method to `/start` and GET method to `/status`
5. Enable CORS for both endpoints
6. Deploy to a stage (e.g., `prod`)

### Step 3: Update HTML Configuration

1. In `index.html`, replace:
   ```javascript
   const API_BASE_URL = 'https://your-api-id.execute-api.region.amazonaws.com/prod';
   ```
   with your actual API Gateway URL
2. Deploy to Netlify or your preferred hosting service

### Step 4: Test the Flow

1. User visits your hosted page
2. Page automatically calls `/start` endpoint
3. Lambda function starts EC2 instance
4. Page polls `/status` endpoint every 5 seconds
5. User plays with the robot-building game while waiting
6. When ready, the page shows "Go to Instance" button
7. User can click to redirect to their EC2 instance

## 🔒 Security Enhancements

For production environments, consider:

1. **Restrict CORS origins**:
   ```python
   headers = {
       'Access-Control-Allow-Origin': 'https://your-app.netlify.app',  # Specific domain
       # ... other headers
   }
   ```

2. **Add API key authentication**:
   ```python
   def lambda_handler(event, context):
       # Check API key
       api_key = event.get('headers', {}).get('x-api-key')
       if api_key != os.environ.get('EXPECTED_API_KEY'):
           return {
               'statusCode': 401,
               'headers': headers,
               'body': json.dumps({'error': 'Unauthorized'})
           }
       # ... rest of your code
   ```

## ✨ Features

This setup provides:
- Real-time status monitoring
- Smooth user experience
- Automatic redirection when ready
- Error handling
- CORS support for browser requests

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