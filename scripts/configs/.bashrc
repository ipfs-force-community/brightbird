export $(ssh-agent | sed -n 1p | grep -o "SSH_AUTH_SOCK=[^;]*")
