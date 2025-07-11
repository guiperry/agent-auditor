// ip-store.js - Module for storing and retrieving the last known working IP address
const fs = require('fs');
const path = require('path');

// Define the path for the IP storage file
// Using /tmp directory which is writable in Netlify Functions
const IP_STORAGE_FILE = '/tmp/last-known-ip.json';

// Function to save IP address and status
async function saveIpAddress(ipAddress, isWorking = true, port = null) {
  try {
    const data = {
      ipAddress,
      isWorking,
      port,
      lastChecked: new Date().toISOString(),
      lastWorking: isWorking ? new Date().toISOString() : null
    };
    
    await fs.promises.writeFile(IP_STORAGE_FILE, JSON.stringify(data, null, 2));
    console.log(`Saved IP address ${ipAddress} to storage file`);
    return true;
  } catch (error) {
    console.error('Error saving IP address:', error);
    return false;
  }
}

// Function to get the last known IP address
async function getLastIpAddress() {
  try {
    // Check if the file exists
    if (!fs.existsSync(IP_STORAGE_FILE)) {
      console.log('No stored IP address found');
      return null;
    }
    
    const data = await fs.promises.readFile(IP_STORAGE_FILE, 'utf8');
    return JSON.parse(data);
  } catch (error) {
    console.error('Error reading IP address:', error);
    return null;
  }
}

// Function to check if a server is responding at the given IP address
async function checkServerAvailability(ipAddress, port = null) {
  const urls = [];
  
  // Add URLs to check based on provided IP and port
  if (port) {
    urls.push(`http://${ipAddress}:${port}`);
  } else {
    // Try standard ports if no specific port provided
    urls.push(`http://${ipAddress}`);       // Standard HTTP port 80
    urls.push(`http://${ipAddress}:8084`);  // Application port
  }
  
  console.log(`Checking server availability at: ${urls.join(', ')}`);
  
  // Try each URL with a fetch request
  for (const url of urls) {
    try {
      // Use node-fetch for server-side fetch
      const fetch = require('node-fetch');
      
      // Set a timeout to avoid waiting too long
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 3000);
      
      const response = await fetch(url, { 
        method: 'HEAD',
        signal: controller.signal
      });
      
      clearTimeout(timeoutId);
      
      // If we get here, the server responded
      console.log(`Server available at ${url}`);
      
      // Extract port from URL if it exists
      const portMatch = url.match(/:(\d+)$/);
      const detectedPort = portMatch ? parseInt(portMatch[1]) : null;
      
      return { 
        available: true, 
        workingUrl: url,
        port: detectedPort
      };
    } catch (error) {
      console.log(`Server not available at ${url}: ${error.message}`);
      // Continue to the next URL
    }
  }
  
  // If we get here, none of the URLs worked
  return { available: false };
}

module.exports = {
  saveIpAddress,
  getLastIpAddress,
  checkServerAvailability
};