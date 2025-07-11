# IP Address Caching Implementation

This document explains the implementation of the IP address caching feature that allows the loader to check if the site is already running at its last known IP address before starting a new GitHub workflow.

## Overview

The implementation adds the ability to:
1. Store the last known working IP address in a file
2. Check if the site is already running at that IP address before triggering a new GitHub workflow
3. Redirect directly to the running site if it's available
4. Update the stored IP address when a new instance is started

## Files Modified/Created

1. **ip-store.js** (New)
   - Core module for storing and retrieving IP addresses
   - Provides functions to save and retrieve IP addresses
   - Includes functionality to check if a server is responding at a given IP

2. **github-start.js** (Modified)
   - Added check for existing running instance before starting workflow
   - Returns "already_running" status if site is already available
   - Includes IP address and redirect URL in response

3. **github-status.js** (Modified)
   - Saves IP address to storage when a successful instance is found
   - Uses the ip-store module to persist IP information

4. **index.html** (Modified)
   - Updated to handle "already_running" status
   - Shows different messages for already running vs newly started instances
   - Provides visual indication that cached IP is being used

5. **package.json** (Modified)
   - Added node-fetch dependency for server-side HTTP requests

## How It Works

1. **Initial Request Flow**:
   - When the page loads, it calls the `github-start` function
   - The function first checks if the site is already running at the last known IP
   - If available, it returns immediately with "already_running" status
   - If not available, it triggers the GitHub workflow as before

2. **IP Storage**:
   - IP addresses are stored in `/tmp/last-known-ip.json`
   - The file includes the IP address, working status, port, and timestamps
   - This location is writable in Netlify Functions environment

3. **Server Availability Check**:
   - The system tries multiple URLs (standard HTTP port and application port)
   - Uses a timeout to avoid waiting too long for responses
   - Returns the first working URL found

4. **User Experience**:
   - User sees a notification that the site is already running
   - Countdown timer shows before automatic redirect
   - User can still choose to go immediately, cancel, or copy the IP

## Benefits

1. **Faster Access**: Users get immediate access to the site if it's already running
2. **Cost Savings**: Avoids unnecessary EC2 instance starts
3. **Improved UX**: Provides clear feedback about using cached IP
4. **Reliability**: Falls back to normal workflow if cached IP is not responding

## Technical Notes

1. **File Storage in Netlify**:
   - Netlify Functions have access to the `/tmp` directory for temporary storage
   - This storage persists across function invocations but may be cleared periodically
   - For more permanent storage, consider using a database or external service

2. **Error Handling**:
   - The implementation includes robust error handling
   - If any part of the IP checking process fails, it falls back to the normal workflow

3. **Timeout Handling**:
   - Server availability checks use a 3-second timeout
   - This prevents long waits if the server is not responding

4. **Port Detection**:
   - The system tries both standard HTTP port (80) and the application port (8084)
   - It remembers which port worked for future checks