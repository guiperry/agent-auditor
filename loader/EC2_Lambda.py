import os
import boto3
import json
import time
from botocore.exceptions import ClientError, WaiterError

# Initialize the EC2 client
REGION = os.environ.get('REGION')
INSTANCE_ID = os.environ.get('INSTANCE_ID')

def lambda_handler(event, context):
    """
    Handles EC2 instance management with API endpoints for the robot loading game
    """
    try:
        # CORS headers for browser requests
        headers = {
            'Access-Control-Allow-Origin': '*',  # Change to your Netlify domain for security
            'Access-Control-Allow-Headers': 'Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token',
            'Access-Control-Allow-Methods': 'GET,POST,OPTIONS',
            'Content-Type': 'application/json'
        }
        
        # Handle preflight OPTIONS requests
        if event.get('httpMethod') == 'OPTIONS':
            return {
                'statusCode': 200,
                'headers': headers,
                'body': json.dumps({'message': 'CORS preflight'})
            }
        
        # Validate environment variables
        if not REGION or not INSTANCE_ID:
            return {
                'statusCode': 500,
                'headers': headers,
                'body': json.dumps({
                    'error': 'Configuration error',
                    'message': 'Missing required environment variables'
                })
            }
        
        ec2 = boto3.client('ec2', region_name=REGION)
        
        # Route handling
        http_method = event.get('httpMethod', 'POST')
        path = event.get('path', '/start')
        
        if http_method == 'POST' and path == '/start':
            return handle_start_instance(ec2, headers)
        elif http_method == 'GET' and path == '/status':
            return handle_check_status(ec2, headers)
        else:
            return {
                'statusCode': 404,
                'headers': headers,
                'body': json.dumps({'error': 'Not found'})
            }
            
    except Exception as e:
        return {
            'statusCode': 500,
            'headers': headers,
            'body': json.dumps({
                'error': 'Internal error',
                'message': str(e)
            })
        }

def handle_start_instance(ec2, headers):
    """Start the EC2 instance"""
    try:
        print(f"🤖 Robot game initiated! Starting EC2 instance: {INSTANCE_ID}")
        
        # Check current instance state
        reservations = ec2.describe_instances(InstanceIds=[INSTANCE_ID])
        instance = reservations['Reservations'][0]['Instances'][0]
        current_state = instance['State']['Name']
        
        if current_state == 'running':
            public_ip = instance.get('PublicIpAddress', 'N/A')
            return {
                'statusCode': 200,
                'headers': headers,
                'body': json.dumps({
                    'status': 'already_running',
                    'message': f'Instance {INSTANCE_ID} is already running',
                    'publicIp': public_ip,
                    'instanceId': INSTANCE_ID
                })
            }
        
        # Start the instance
        ec2.start_instances(InstanceIds=[INSTANCE_ID])
        print(f"✅ Instance {INSTANCE_ID} start command sent")
        
        return {
            'statusCode': 200,
            'headers': headers,
            'body': json.dumps({
                'status': 'starting',
                'message': f'Instance {INSTANCE_ID} is starting up',
                'instanceId': INSTANCE_ID,
                'estimatedTime': '2-3 minutes'
            })
        }
        
    except ClientError as e:
        error_code = e.response['Error']['Code']
        error_message = e.response['Error']['Message']
        print(f"❌ AWS API error: {error_code} - {error_message}")
        return {
            'statusCode': 500,
            'headers': headers,
            'body': json.dumps({
                'error': error_code,
                'message': error_message
            })
        }

def handle_check_status(ec2, headers):
    """Check the current status of the EC2 instance"""
    try:
        # Fallback IP address to use if instance IP is not available
        FALLBACK_IP = "3.146.37.27:8080"
        
        reservations = ec2.describe_instances(InstanceIds=[INSTANCE_ID])
        instance = reservations['Reservations'][0]['Instances'][0]
        
        current_state = instance['State']['Name']
        public_ip = instance.get('PublicIpAddress', 'N/A')
        
        # Determine completion status
        is_ready = current_state == 'running'
        
        response_data = {
            'instanceId': INSTANCE_ID,
            'state': current_state,
            'publicIp': public_ip if is_ready else None,
            'isReady': is_ready,
            'message': f'Instance is {current_state}'
        }
        
        if is_ready:
           # Use fallback IP if public_ip is not available
            if public_ip != 'N/A':
                response_data['redirectUrl'] = f'http://{public_ip}'
                print(f"🎉 Instance {INSTANCE_ID} is ready! Public IP: {public_ip}")
            else:
                response_data['redirectUrl'] = f'http://{FALLBACK_IP}'
                print(f"🎉 Instance {INSTANCE_ID} is ready! Using fallback IP: {FALLBACK_IP}")
        else:
            print(f"⏳ Instance {INSTANCE_ID} status: {current_state}")
        
        return {
            'statusCode': 200,
            'headers': headers,
            'body': json.dumps(response_data)
        }
        
    except ClientError as e:
        error_code = e.response['Error']['Code']
        error_message = e.response['Error']['Message']
        return {
            'statusCode': 500,
            'headers': headers,
            'body': json.dumps({
                'error': error_code,
                'message': error_message
            })
        }