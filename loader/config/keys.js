// config/keys.js
// This file loads AWS credentials from environment variables or uses defaults for development

// For production, we'll use Netlify's environment variables with different names
// to avoid the reserved environment variable restrictions
if (process.env.NODE_ENV === 'production') {
  module.exports = {
    region: process.env.NETLIFY_AWS_REGION || 'us-east-1',
    accessKeyId: process.env.NETLIFY_AWS_KEY_ID,
    secretAccessKey: process.env.NETLIFY_AWS_SECRET_KEY,
    instanceId: process.env.NETLIFY_EC2_INSTANCE_ID
  };
} else {
  // For local development, we can use the standard AWS environment variables
  // or hardcoded development values (not recommended for real credentials)
  module.exports = {
    region: process.env.AWS_REGION || 'us-east-1',
    accessKeyId: process.env.AWS_ACCESS_KEY_ID || 'development-key-id',
    secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY || 'development-secret-key',
    instanceId: process.env.EC2_INSTANCE_ID || 'i-development'
  };
}