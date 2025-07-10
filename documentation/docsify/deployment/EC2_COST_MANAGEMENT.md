# EC2 Instance Cost Management

This document explains how to use the EC2 instance cost management features in the Agent Auditor deployment process.

## Overview

The Agent Auditor deployment system now includes automatic EC2 instance management to help reduce AWS costs. The system can:

1. Automatically start your EC2 instance before deployment
2. Deploy the application
3. Automatically stop the EC2 instance after deployment (whether successful or failed)

This approach ensures that your EC2 instances only run when needed, significantly reducing your AWS costs.

## How It Works

The deployment process uses the following workflow:

1. **Find Instance**: The playbook uses the `amazon.aws.ec2_instance_info` module to find your EC2 instance by its Name tag.
2. **Start Instance**: It then starts the instance and waits for it to be fully operational.
3. **Deploy Application**: Once the instance is running, it deploys the application using the existing agent_auditor role.
4. **Stop Instance**: After deployment (whether successful or not), it stops the instance to save costs.

## Configuration

The EC2 instance management features can be configured in the Ansible defaults file:

```yaml
# EC2 Instance Management Settings
ec2_instance_tag_name: "agent-auditor"  # The Name tag of your EC2 instance
ec2_wait_timeout: 300                   # Timeout in seconds to wait for instance operations
ec2_auto_start: true                    # Whether to automatically start the instance before deployment
ec2_auto_stop: true                     # Whether to automatically stop the instance after deployment
ec2_ssh_user: "ubuntu"                  # SSH user for connecting to the EC2 instance
ec2_ssh_key_file: "~/.ssh/AEGONG.pem"  # Path to the SSH private key file
```

You can override these settings in your group variables or by using extra vars when running the playbook.

## Prerequisites

1. AWS CLI configured with appropriate credentials
2. Ansible installed with the Amazon AWS collection:
   ```bash
   ansible-galaxy collection install amazon.aws
   ```
3. An existing EC2 instance with the tag "Name:agent-auditor" (or modify the tag in the configuration)

## Usage

The EC2 instance management is integrated with the existing deployment process. Simply use the standard deployment command:

```bash
make deploy
```

The deployment process will automatically:
1. Start your EC2 instance if it's stopped
2. Wait for the instance to be ready
3. Deploy the application to the dynamic IP address of the instance
4. Stop the instance after deployment

### Deployment Behavior

The playbook has two deployment strategies:

1. **Dynamic IP Deployment** (when `ec2_auto_start: true`):
   - The playbook starts the EC2 instance
   - It retrieves the current public IP address
   - It deploys to this dynamic IP address
   - This ensures deployment works even if the instance's IP changes

2. **Static IP Deployment** (when `ec2_auto_start: false`):
   - The playbook uses the static inventory file (`ansible/inventory/hosts.ini`)
   - It deploys to the IP addresses defined in the inventory
   - Use this mode if you have a fixed IP or Elastic IP

## Troubleshooting

If you encounter issues with the EC2 instance management:

1. **Instance Not Found**: Ensure your EC2 instance has the correct Name tag (default: "AEGONG-Server")
2. **Permission Issues**: Verify that your AWS credentials have the necessary permissions to start/stop EC2 instances
3. **Timeout Errors**: If your instance takes longer to start, increase the `ec2_wait_timeout` value
4. **SSH Connection Issues**: Check that your SSH key path is correct and the security group allows SSH access

## Disabling Automatic Instance Management

If you need to disable the automatic instance management:

1. **Disable Both Start and Stop**:
   ```yaml
   ec2_auto_start: false
   ec2_auto_stop: false
   ```

2. **Disable Only Auto-Stop** (useful for debugging):
   ```yaml
   ec2_auto_start: true
   ec2_auto_stop: false
   ```

## Security Considerations

- Configure the SSH key path (`ec2_ssh_key_file`) to match your key location
- Consider using AWS Secrets Manager or Ansible Vault for sensitive information
- Restrict SSH access to specific IP addresses in your security group configuration
- Use a dedicated IAM role with minimal permissions for the EC2 instance