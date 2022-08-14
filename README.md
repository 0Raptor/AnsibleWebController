# Ansible Web Controller (AWC)

AWC is a web application written in GO that allows you to execute preconfigured tasks as a user of your Linux system. The idea is to run Ansible Playbooks and view the results remotely. Arguments can be supplied using HTML-Forms. The forms to fill in and commands to execut will be generated dynamically based on the `commands.xml` file. Beside Playbooks technically any Linux command can be configured as a task.

This application was created for people who want to run scripts remotely without having to access the shell. Especially, for Ansible users who do not need the vast amount of functions included in AWX or Ansible Tower.

## Content

- [Ansible Web Controller (AWC)](#ansible-web-controller-awc)
  - [Content](#content)
  - [Overview](#overview)
  - [Annotations](#annotations)
  - [Setup](#setup)
    - [Environment](#environment)
    - [Installation](#installation)
      - [Use Ansible-Vault insted of sudoers](#use-ansible-vault-insted-of-sudoers)
    - [Configuration](#configuration)
      - [Managing Settings](#managing-settings)
      - [Defining Tasks](#defining-tasks)
  - [API Usage](#api-usage)
  - [Upgrading](#upgrading)
  - [Support Me](#support-me)
  - [Legal Notice](#legal-notice)
  - [License](#license)

## Overview

**Language**  
GO, HTML

**OS**  
Linux (tested on CentOS 8; any distribution should work)

**Abstract**  
Define Shell/ Bash commands (and parameter if required); Show commands on webinterface; Enter parameter (if defined) in generated web-form; Run commands (automatically replace parameter-spacers with user input and write to logfile) from web; List and read log files (result of commands) in webinterface

**Dependencies**  
Using Binary: None (to run the bare application)  
Compiling: [GO](https://go.dev/)

## Annotations

This application is NOT designed for production use! You may use it for testing or in home lab environments.  
DO NOT run this application as root! Once configured the commands can be executed with any parameter. You have to make sure that the parameter will not endanger your system. I recommend using a non root user that can only connect to the Ansible clients via SSH.  
Currently there is no authentication included in this application. I recommend blocking the application's port via a firewall and only allow access to the application from localhost or a reverse proxy with authentication (see Installation).

Following the setup instructions, you will not be able to make system-level changes (e.g. updates) to the server that runs Ansible itself. This is a security measure, but you could grant more privileges to the user running the web-application at your own risk.

The included HTML UI/ front-end is (to be kind) very simple. If the UI does not appeal to you and you are a better front-end developer then I am, I designed this application so you can change the complete UI without having to touch or recompile the binary. Just change the files inside the HTML directory to your needs!

## Setup

The **Environment** chapter explains the intended use case of the application. If you do not use it for executing Ansible Playbooks you can skip this part and just create a non-admin user that will be used to execute the application.  
The commands in the **Installation** chapter are tested on Centos 8 Stream. If you are using another OS you may skip the SELinux and firewall commands and use `apache2` instead of `https`. Beware, the directories for the webserver will only match if you are using `httpd`!  
The **Configuration** explains how to use the configuration files of this application to configure your remotely executable commands. Read this chapter carefully!

### Environment

I started with a fresh installed Centos 8 Stream Server without desktop. My account is not root, but part of the sudo group.

I wanted to setup Ansible as management tool and *AWC* to remotely execute my Playbooks.

1. Install Ansible

```Bash
sudo yum install epel-release
sudo yum install ansible
```

2. Create user to run Playbooks (NOT part of sudo-group) and deny ssh login

```Bash
sudo useradd ansible -m
sudo passwd ansible

echo "DenyUsers ansible" >> /etc/ssh/sshd_config
```

3. Setup demo Ansible files

```Bash
sudo su - ansible
mkdir inventory
mkdir playbooks
nano inventory/hosts
    # add your servers that will be managed - compare 1st content below
nano playbooks/create-testfile.yml
    # add 2nd content below
nano playbooks/update-ubuntu.yml
    # add 3rd content below
```

```None
[ubuntu]
server1.example.com
server2.example.com
10.0.0.6

[centos]
localhost
```

```YML
---

- hosts: "{{ targets }}"
  become: yes # become root user on target machine
  gather_facts: no # default this will be set to yes - collect data before starting

  #vars:
    #ansible_port: 12345 # add this lines if you are not using the default ssh port (22)

  tasks:
  - name: Creating a test file with content
    copy:
      dest: "{{ path }}"
      content: "{{ content }}"
```

```YML
---
# works for ubuntu & debian
- hosts: ubuntu
  become: yes
  gather_facts: no

  tasks:
    - name: Apt Update
      apt: update_cache=yes force_apt_get=yes cache_valid_time=3600

    - name: Apt Upgrade
      apt: upgrade=dist force_apt_get=yes

    - name: Check if reboot is required
      register: reboot_required_file
      stat: path=/var/run/reboot-required get_md5=no

    - name: Reboot if required (e.g. due to new kernel)
      reboot:
        msg: "Reboot initiated by Ansible due to kernel updates"
        connect_timeout: 5
        reboot_timeout: 300
        pre_reboot_delay: 0
        post_reboot_delay: 30
        test_command: uptime
      when: reboot_required_file.stat.exists
```

4. Add a manged node (on the node)

Create a user that Ansible will connect to on the managed nodes. This user will be part of the sudo group so Ansible will be able to execute tasks that require root privileges. The user will be configured so it can use `sudo` without entering a password (so that no password has to be provided to Ansible).

> **HINT**  
> Instead of allowing the user to use "sudo" without entering its password, you might supply the password to Ansible. Do not store it in plaintext on your Ansible controller - use `ansible-vault`!  
> [Click here to get instructions]](#use-ansible-vault-insted-of-sudoers)

```Bash
sudo su -

# create user "ansible"
useradd ansible -m
passwd ansible

# add user to sudo-group --> grant right to user "sudo ..."
sudo usermod -aG sudo ansible
# allow user to use "sudo" without entering its password
echo "ansible ALL=(ALL:ALL) NOPASSWD:ALL" >> /etc/sudoers

# disable ssh login to user
echo "Match User ansible" >> /etc/ssh/sshd_config
echo "        PasswordAuthentication no" >> /etc/ssh/sshd_config

# STOP! complete step 5 first - then enter the last two commands on the managed node

# restart sshd to apply changes
systemctl restart sshd

exit
```

5. Add a manged node (on the server with Ansible installed)

```Bash
sudo su - ansible

# WARN - Only run ssh-keygen command the first time you set up a managed node (DO NOT RUN IT IF YOU ALREADY HAVE A PRIVATE SSH KEY)
# generate private ssh key
ssh-keygen -t ecdsa -b 521
    # just hit enter without providing any text to questions

# copy public ssh key to managed node so Ansible can log in with its private key
ssh-copy-id ansible@<FQDN> #-p 12345 # add the -p argument when using a custom port
    # you will be prompted for a password --> use the one set up in step 4
```

### Installation

1. Download binary (replace url with [latest version](https://git.0raptor.earth/Raptor/AnsibleWebController/releases)) and set permission

```Bash
# create folders in user's folder
sudo mkdir -p /home/ansible/awc
sudo mkdir -p /home/ansible/logs
# prepare log counter
echo "0" > /home/ansible/logs/cntr
# download the binary and unzip it into the created folders
curl -O https://git.0raptor.earth/attachments/40809c62-b93b-4156-bb0d-a84329ed3c07
sudo unzip AWC-Linux_AMD64.zip -d /home/ansible/awc
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

8. Secure webserver with HTTPS

```Shell
sudo su -

# install ssl mod
yum install mod_ssl -y

# create directory for certificates and import them
mkdir /etc/httpd/ssl
openssl dhparam -out /etc/httpd/ssl/dhparam.pem 4096 # generate dhparam
mv awc.key /etc/httpd/ssl/ # import website's private key
mv awc.crt /etc/httpd/ssl/ # import website's certificate
mv intermediate.crt /etc/httpd/ssl/ # import intermediate/ issuer certificate

# remove read access on certs for non root users
chmod 660 /etc/httpd/ssl/*

# allow httpd to read it
semanage fcontext -a -t httpd_sys_content_t "/etc/httpd/ssl(/.*)?"
restorecon -Rv "/etc/httpd/ssl"

# update httpd.conf
echo "Include conf/httpd-ssl.conf" >> /etc/httpd/conf/httpd.conf
nano /etc/httpd/conf/httpd-ssl.conf
    # add 1st content below

# update website confog
nano /etc/httpd/conf.d/awc.conf
    # Change file to 2nd content bellow

# restart webserver
systemctl restart httpd

exit
```

```None
SSLCipherSuite EECDH+AESGCM:EDH+AESGCM:AES256+EECDH:AES256+EDH
SSLProtocol All -SSLv2 -SSLv3 -TLSv1 -TLSv1.1
SSLHonorCipherOrder On
Header always set X-Frame-Options DENY
Header always set X-Content-Type-Options nosniff
SSLCompression off
SSLUseStapling on
SSLStaplingCache "shmcb:logs/stapling-cache(150000)"
SSLSessionTickets Off
SSLOpenSSLConfCmd DHParameters "/etc/httpd/ssl/dhparam.pem"
```

```None
<VirtualHost *:80>
    ServerName <YOUR DOMAIN NAME>
    ServerAlias <YOUR ALIAS DOMAIN NAME>

    Redirect permanent / https://<YOUR DOMAIN NAME>/
</virtualhost>
<VirtualHost *:443>
    ProxyPreserveHost On

    ProxyPass / http://localhost:8080/
    ProxyPassReverse / http://localhost:8080/

    Timeout 5400
    ProxyTimeout 5400

    ServerName <YOUR DOMAIN NAME>
    ServerAlias <YOUR ALIAS DOMAIN NAME>

    SSLEngine on
    SSLCertificateFile "/etc/httpd/ssl/awc.crt"
    SSLCertificateKeyFile "/etc/httpd/ssl/awc.key"
    SSLCertificateChainFile "/etc/httpd/ssl/intermediate.crt"

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

#### Use Ansible-Vault insted of sudoers

If you feel uncomfortable allowing a user to use root-privileges without entering a password, you can use the following commands on the server with Ansible installed instead of performing "echo "ansible ALL=(ALL:ALL) NOPASSWD:ALL" >> /etc/sudoers" on each node.

The password for the ansible-user on the nodes will be stored encrypted on the server and will be read and send to the node by Ansible when it is running a playbook on the node.

```Bash
cd /home/ansible
mkdir vars
cd vars
head -c 250 /dev/random | base64 | head -c 250 > vault.pass # generate random password

# create vault using the password file
ansible-vault create vault.yml --vault-password-file vault.pass
  # file will be opened in vi-editor
  # tap i to get into insert mode and add the content below
  # tap ESC (to stop editing), type :w and hit enter (to save) and type :q and hit enter (to quit)
```

```YML
ansible_port: 12345 # you can also specify your custom-ssh-port here once instead in each playbook
ansible_sudo_pass: "YOUR PASSWORD FOR THE ANSIBLE USER ON THE MANAGED NODES"
```

You can edit the contents of your vault at any time using `ansible-vault edit vault.yml --vault-password-file vault.pass`. If you sync your playbooks via git, you should consider adding `vault.pass` to your `.gitignore`-file so not everybody can read your passwords.

To use the variables in your vault you have to include it in every playbook, using the following code before defining your tasks. Furthermore, you have to call the playbook with an additional argument containing the path to your password-file (e.g. `ansible-playbook create-testfile.yml --vault-password-file ../vars/vault.pass -i ../inventory/hosts`).

```YML
vars_files:
  - ../vars/vault.yml

tasks:
  - [...]
```

### Configuration

This application is highly customizable. You can (almost) the complete GUI without recompiling the code. Moreover, you can define variables for your tasks which will be prompted on the webserver. This section will only cover how to change the [settings](#managing-settings) and how to [define tasks](#defining-tasks).

#### Managing Settings

*A default configuration is shipped with the binary. It's located [here](/config/settings.xml).*

- **port**
  - Specifies the port the application will listen on (default: ":8080")
  - DO NOT forget the colon (:)
- **logdir**
  - Directory the application will search for logs and create the files containing executed commands' outputs
  - The user running the application has to have READ and WRITE permission to this directory

#### Defining Tasks

*A default configuration, including two tasks and utilizing the playbooks created above, is shipped with the binary. It's located [here](/config/commands.xml).*

- Each task will be defined between `<task> [...] </task>` nested inside the outer `xml` container
- Each task MUST have a `name` and `command` container
  - **name**
    - Name of the task shown on the web page (so the user will know what it does)
  - **command**
    - The command that will be executed when the task is run from the webpage
    - You can use variables as described below (see **var**)
    - The assembled command will be executed inside a bash: `bash -c '<COMMAND> > <LOGFILE>'`
      - All results will be written into the log file and can be accessed via the web interface.
  - You can add a `form` container if user inputs are required
  - You can omit the `form` container BUT you MUST NOT add any container beside these three (otherwise the task will be ignored)
- The `form` container must have at least one `input` container, but may have multiple
  - Each `input` container represents a variable you can use in the command and will generate an html-input on the web page
- Each `input` container MUST have a `type`, `label` and `var` container
  - **type**
    - Specifies the type of input that will be displayed on the web page
    - Options are
      - *text*
        - Requests a plain text input
      - *password*
        - Like a normal text, but the input is \*\*\*-out on the web page
      - *checkbox*
        - The user may check the box: checked --> true
      - *dropdown*
        - Creates a drop-down list the user can select an item from
        - You have to specify the available items in an `options` container: Its nested inside `input`, too
          - Options are semicolon (;) seperated: e.g. `ubuntu;centos;all`
          - This only applies when having an input with `type=dropdown` - otherwise DO NOT add an `options` container
  - **label**
    - A text that will be displayed in front of the input to explain the user which information he has to provide on the web page
  - **var**
    - Name the input will have
      - ONLY USE alphanumeric letters, DO NOT use spaces
      - The name must be UNIQUE inside one `form` container
        - In the best case it will be unique throughout all tasks, but that's not mandatory
    - The **command** can use the **var**-name as spacer: The spacer (`??<VAR>??`) will be replaced with the html form's user input
      - e.g. `[...] --extra-vars "path=??path?? content=??content?? targets=??hosts??"`
        - The task will have two inputs: one with `var=path` and a second with `var=content`
      - It is useless to define an input if the **var** is not used inside the **command**

## API Usage

This application is designed to be used by a human operating on the webpage. Nevertheless, tasks can be executed and the configuration can be reloaded through API requests.

- Execute a task
  - `http[s]://<YOUR DOMAIN NAME>/run?id=<TASK ID>[&<VAR>=<VALUE>]*`
  - Tasks are executed via the `/run` URL
    - You have to deliver at least the *id*
      - The tasks in `commands.xml` get their IDs from top to button starting with 0
      - The first task has id 0, the second id 1 and so on
    - If your task has inputs you have to deliver values to them
      - Each value is delivered starting with a `&` followed by the string specified in the `var` container an `=` and the value you want to assign to the variable
      - Make sure to supply all inputs - The API will NOT validate your data
- Reload Configuration
  - `http[s]://<YOUR DOMAIN NAME>/refresh`
  - This will apply changes in `settings.xml` (only *logdir*) and `commands.xml`
  - A changed port can only be applied by restarting the service

Remember to supply your credentials when using the reverse proxy with *HTTP Basic Authentication*.

## Upgrading

After a new release you can replace the binary (`awc`) and html-files with the new binary and html-files. I am trying to ensure that everything around it (logs' directory, configuration files and service) will be compatible.  
When using SELinux you have to relabel the binary after an upgrade using `sudo restorecon -v /home/ansible/awc/awc`!

## Support Me

You like my application? I would be thankful for a small donation to finance my infrastructure (or the next cup of coffee).
[Donate via "Buy Me A Coffee"](https://www.buymeacoffee.com/0Raptor)

## Legal Notice

This project has legally nothing to do with Red Hat's projects [Ansible](https://www.ansible.com/), [AWX](https://www.ansible.com/community/awx-project) or [Automation controller](https://www.ansible.com/products/controller) nor uses any of its source code. These projects, products and names belong to [Red Hat Inc.](https://www.redhat.com/en).

This project was designed to remotely execute Ansible Playbooks which is represented in the name.  
But it can also be used to execute any other kind of script and application.

## License

This application is published undes GNU GNU GENERAL PUBLIC LICENSE Version 3 as refered in the LICENSE-File.

Copyright (C) 2022 0Raptor

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not, see [www.gnu.org/licenses/](https://www.gnu.org/licenses/).
