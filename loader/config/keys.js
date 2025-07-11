// config/keys.js
// This file loads AWS credentials from environment variables or uses defaults for development

// Helper function to clean environment variables
// Removes any whitespace, quotes, or other unwanted characters
function cleanEnvVar(value) {
  if (!value) return value;
  
  // Remove leading/trailing whitespace
  let cleaned = value.trim();
  
  // Remove surrounding quotes if present
  if ((cleaned.startsWith('"') && cleaned.endsWith('"')) || 
      (cleaned.startsWith("'") && cleaned.endsWith("'"))) {
    cleaned = cleaned.substring(1, cleaned.length - 1);
  }
  
  return cleaned;
}

// Log all available environment variables for debugging (names only)
console.log('Available environment variables (names only):');
Object.keys(process.env)
  .filter(key => key.includes('AWS') || key.includes('NETLIFY') || key === 'NODE_ENV' || key === 'CONTEXT')
  .forEach(key => console.log(`  - ${key}`));

// Check for Netlify environment first, regardless of NODE_ENV
// This ensures we always try to use Netlify variables when deployed
if (process.env.NETLIFY || process.env.CONTEXT) {
  console.log('Running in Netlify environment, using Netlify environment variables');
  
  // Get and clean environment variables
  const region = cleanEnvVar(process.env.NETLIFY_AWS_REGION) || 'us-east-2';
  let accessKeyId = cleanEnvVar(process.env.NETLIFY_AWS_KEY_ID);
  let secretAccessKey = cleanEnvVar(process.env.NETLIFY_AWS_SECRET_KEY);
  let instanceId = cleanEnvVar(process.env.NETLIFY_EC2_INSTANCE_ID);
  
  // If Netlify-specific variables aren't set, try standard AWS variables as fallback
  if (!accessKeyId || !secretAccessKey) {
    console.log('Netlify-specific AWS credentials not found, trying standard AWS environment variables as fallback');
    accessKeyId = cleanEnvVar(process.env.AWS_ACCESS_KEY_ID);
    secretAccessKey = cleanEnvVar(process.env.AWS_SECRET_ACCESS_KEY);
    
    if (!instanceId) {
      instanceId = cleanEnvVar(process.env.EC2_INSTANCE_ID);
    }
  }
  
  // Log partial credentials for debugging (first 4 chars only)
  console.log(`Region: ${region}`);
  console.log(`Instance ID: ${instanceId || 'Not provided'}`);
  console.log(`Access Key ID: ${accessKeyId ? accessKeyId.substring(0, 4) + '...' : 'Not provided'}`);
  console.log(`Secret Access Key: ${secretAccessKey ? secretAccessKey.substring(0, 4) + '...' : 'Not provided'}`);
  
  module.exports = {
    region,
    accessKeyId,
    secretAccessKey,
    instanceId
  };
} else if (process.env.NODE_ENV === 'production') {
  console.log('Running in production environment, using production environment variables');
  
  // Get and clean environment variables
  const region = cleanEnvVar(process.env.AWS_REGION) || 'us-east-2';
  const accessKeyId = cleanEnvVar(process.env.AWS_ACCESS_KEY_ID);
  const secretAccessKey = cleanEnvVar(process.env.AWS_SECRET_ACCESS_KEY);
  const instanceId = cleanEnvVar(process.env.EC2_INSTANCE_ID);
  
  module.exports = {
    region,
    accessKeyId,
    secretAccessKey,
    instanceId
  };
} else {
  // For local development, we can use the standard AWS environment variables
  // or hardcoded development values (not recommended for real credentials)
  console.log('Running in development environment, using development environment variables');
  
  // Get and clean environment variables
  const region = cleanEnvVar(process.env.AWS_REGION) || 'us-east-2';
  const accessKeyId = cleanEnvVar(process.env.AWS_ACCESS_KEY_ID) || 'development-key-id';
  const secretAccessKey = cleanEnvVar(process.env.AWS_SECRET_ACCESS_KEY) || 'development-secret-key';
  const instanceId = cleanEnvVar(process.env.EC2_INSTANCE_ID) || 'i-development';
  
  module.exports = {
    region,
    accessKeyId,
    secretAccessKey,
    instanceId
  };
}