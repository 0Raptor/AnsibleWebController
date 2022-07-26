# Ansible Web Controller (AWC)

AWC is a web application written in GO that allows you to execute preconfigured tasks as a user of your Linux system. The idea is to run Ansible Playbooks and view the results remotely. Arguments can be supplied using HTML-Forms. The forms to fill in and commands to execut will be generated dynamically based on the `commands.xml` file. Beside Playbooks technically any Linux command can be configured as a task.

This application was created for people who want to run scripts remotely without having to access the shell. Especially, for Ansible users who do not need the vast amount of functions included in AWX or Ansible Tower.

## Annotations

This application is NOT designed for production use! You may use it for testing or in home lab environments.  
DO NOT run this application as root! Once configured the commands can be executed with any parameter. You have to make sure that the parameter will not endanger your system. I recommend using a non root user that can only connect to the Ansible clients via SSH.  
Currently there is no authentication included in this application. I recommend blocking the application's port via a firewall and only allow access to the application from localhost or a reverse proxy with authentication (see Installation).

## Setup

The **Environment** chapter explains the intended use case of the application. If you do not use it for executing Ansible Playbooks you can skip this part and just create a non-admin user that will be used to execute the application.  
The commands in the **Installation** chapter are tested on Centos 8 Stream. If you are using another OS you may skip the SELinux and firewall commands and use `apache2` instead of `https`. Beware, the directories for the webserver will only match if you are using `httpd`!  
The **Configuration** explains how to use the configuration files of this application to configure your remotely executable commands. Read this chapter carefully!

### Environment

**TODO**

### Installation

1. Download binary and set permission

```Bash
# create folders in user's folder
sudo mkdir -p /home/ansible/awc
sudo mkdir -p /home/ansible/logs
# download the binary and unzip it into the created folders
curl <URL>
sudo unzip <FILE> -d /home/ansible/awc
# set permissions to let user manage the directories and execute the binary
sudo chmod +x /home/ansible/awc/awc
sudo chown -R ansible:ansible /home/ansible/awc
sudo chown -R ansible:ansible /home/ansible/logs
```

2. Install webserver to use as reverse proxy with authentication

```Bash
# use package manager to install apache2 webserver
sudo yum install httpd -y
# create a new password file with your username (DO NOT USE -c option when adding more users)
sudo htpasswd -c /etc/httpd/.htpasswd <USERNAME>
# update firewall to allow HTTP (80/TCP) and HTTPS (443/TCP)
firewall-cmd --add-service https --permanent
firewall-cmd --add-service http --permanent
firewall-cmd --reload
# setup webserver to start during boot and start it the first time
sudo systemctl enable httpd
sudo sytemctl start httpd
```

3. Open your server's domain name/ IP in your browser to confirm the HTTPD-Testpage is shown
4. Configure the reverse proxy

```Bash
nano /etc/httpd/conf.d/welcome.conf
    # Comment evrything out
nano /etc/httpd/conf.d/awc.conf
    # Add content bellow

# allow webserver to connect to other webservices - required so it can communicate with the webserver delivered by the binary
sudo setsebool -P httpd_can_network_connect on
# restart webserver to apply changes
sudo sytemctl restart httpd
```

```None
<VirtualHost *:80>
    ProxyPreserveHost On

    ProxyPass / http://localhost:8080/
    ProxyPassReverse / http://localhost:8080/

    Timeout 5400
    ProxyTimeout 5400

    ServerName <YOUR DOMAIN NAME>
    ServerAlias <YOUR ALIAS DOMAIN NAME>

    <Proxy *>
        Order deny,allow
        Allow from all
        Authtype Basic
        Authname "Password Required"
        AuthUserFile /etc/httpd/.htpasswd
        Require valid-user
    </Proxy>
</virtualhost>
```

5. Start *AWC* as the non-root-user

```Shell
sudo su - ansible
cd /home/ansible/awc
./awc
```

6. Refresh your browser (compare 2.) and enter credentials --> *AWC* interface should be visible
7. Make *AWC* start during boot

```Shell
# Allow systemd execution in selinux
sudo semanage fcontext -a -t bin_t "/home/ansible/awc/awc"
sudo restorecon -v /home/ansible/awc/awc 
# Create service
sudo nano /etc/systemd/system/awc.service
    # Add content bellow
# Load, start and enable service
sudo systemctl daemon-reload
sudo systemctl start awc
sudo systemctl status awc # should say "active (running)"
sudo systemctl enable awc
```

```None
[Unit]
Description=Asible Web Controller Service
After=network.target

[Service]
Type=simple
User=ansible
Group=ansible
WorkingDirectory=/home/ansible/awc
ExecStart=/home/ansible/awc/awc
Restart=always


[Install]
WantedBy=multi-user.target
```

### Configuration

**TODO**

## Upgrading

After a new release you can replace the binary (`awc`) and html-files with the new binary and html-files. I am trying to ensure that everything around it (logs' directory, configuration files and service) will be compatible.  
When using SELinux you have to relabel the binary after an upgrade using `sudo restorecon -v /home/ansible/awc/awc`!

## License

This application is published undes GNU GNU GENERAL PUBLIC LICENSE Version 3 as refered in the LICENSE-File.

Copyright (C) 2022 0Raptor

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not, see [www.gnu.org/licenses/](https://www.gnu.org/licenses/).
