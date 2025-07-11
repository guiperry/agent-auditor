// config/keys.js
// This file loads AWS credentials from environment variables or uses defaults for development

// Check for Netlify environment first, regardless of NODE_ENV
// This ensures we always try to use Netlify variables when deployed
if (process.env.NETLIFY || process.env.CONTEXT) {
  console.log('Running in Netlify environment, using Netlify environment variables');
  module.exports = {
    region: process.env.NETLIFY_AWS_REGION || 'us-east-2',
    accessKeyId: process.env.NETLIFY_AWS_KEY_ID,
    secretAccessKey: process.env.NETLIFY_AWS_SECRET_KEY,
    instanceId: process.env.NETLIFY_EC2_INSTANCE_ID
  };
} else if (process.env.NODE_ENV === 'production') {
  console.log('Running in production environment, using production environment variables');
  module.exports = {
    region: process.env.AWS_REGION || 'us-east-2',
    accessKeyId: process.env.AWS_ACCESS_KEY_ID,
    secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY,
    instanceId: process.env.EC2_INSTANCE_ID
  };
} else {
  // For local development, we can use the standard AWS environment variables
  // or hardcoded development values (not recommended for real credentials)
  console.log('Running in development environment, using development environment variables');
  module.exports = {
    region: process.env.AWS_REGION || 'us-east-2',
    accessKeyId: process.env.AWS_ACCESS_KEY_ID || 'development-key-id',
    secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY || 'development-secret-key',
    instanceId: process.env.EC2_INSTANCE_ID || 'i-development'
  };
}