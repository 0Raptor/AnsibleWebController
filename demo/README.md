# Ansible Web Controller - Demo Files

This directory includes some Ansible-Playbooks and their matching configuration for AWC in this README.  
You can use this information as a base to create your own set of playbooks or even use them in your system.

- [Ansible Web Controller - Demo Files](#ansible-web-controller---demo-files)
  - [Update Debian & Ubuntu Hosts (hosts using apt)](#update-debian--ubuntu-hosts-hosts-using-apt)
  - [Install custom .deb (copy to host or from url)](#install-custom-deb-copy-to-host-or-from-url)
  - [Create Test File on Hosts](#create-test-file-on-hosts)
  - [Configure New System](#configure-new-system)

All Playbooks are set up to use Ansible-Vault.  
You have to add the XML-code below between the outer xml-brakets (`<xml> CODE GOES HERE </xml>`) in the `commands.xml`.

## Update Debian & Ubuntu Hosts (hosts using apt)

Updates all managed Debian- and Ubuntu hosts using `apt`. Use can decide whether to use a normal `upgrade` or a `dist-upgrade` and if the host may reboot if required.

[View Playbook](UpdateApt.yml)

Requires:

- inventory-file (with a list of Debian- and Ubuntu hosts)
- vault password

Inputs:

- distupgrade - Use `apt dist-upgrade` when `on`. Otherwise, `apt update`
- allowrestart - Restarts the host at the end of the script if an installed update requested it when `on`

```XML
<task>
    <name>Update Debian- and Ubuntu Hosts</name>
    <command>ansible-playbook /home/ansible/ansible/playbooks/UpdateApt.yml -i /home/ansible/ansible/inventory/hosts --vault-password-file /home/ansible/ansible/vars/vault.pass --extra-vars "allowrestart=??restart?? distupgrade=??distupgrade??"</command>
    <form>
        <input>
            <type>checkbox</type>
            <label>Use dist-upgrade</label>
            <var>restart</var>
        </input>
        <input>
            <type>checkbox</type>
            <label>Rebbot system if required</label>
            <var>restart</var>
        </input>
    </form>
</task>
```

*The `var` name in the xml-file does not have to match its name in the Playbook. You have to assign them in the `command`'s `--extra-vars "[...]"`-section (compare `allowrestart=??restart??` above).*

## Install custom .deb (copy to host or from url)

Installs a .deb file on specified hosts or hostgroups using `apt`. The file can be copied from Ansible-Controller-Host or specified as a URL.

[View Playbook](InstallCustomDeb.yml)

Requires:

- inventory-file
- vault password

Inputs:

- path - Full file-path or URL of the .deb
- isurl - `on` when *path* is an URL
- targets - List (comma-separated without spaces) of host-ips, host-names, hostgroup-names or just one of these values) to deploy the .deb to

```XML
<task>
    <name>Install .deb (APT)</name>
    <command>ansible-playbook /home/ansible/ansible/playbooks/InstallCustomDeb.yml -i /home/ansible/ansible/inventory/hosts --vault-password-file /home/ansible/ansible/vars/vault.pass --extra-vars "path=??path?? isurl=??isurl?? targets=??hosts??,"</command>
    <form>
        <input>
            <type>text</type>
            <label>Filepath (on Ansible-Controller) or URL</label>
            <var>path</var>
        </input>
        <input>
            <type>checkbox</type>
            <label>Path is URL</label>
            <var>isurl</var>
        </input>
        <input>
            <type>text</type>
            <label>Hosts (IPs, Hostgroups: comma-separated-list without spaces)</label>
            <var>hosts</var>
        </input>
    </form>
</task>
```

*If you want the option to just pass ONE host or hostgroup into a variable that will be used to define `- hosts: "{{ targets }}"` you have to always add a comma to the end of the string (so Ansible thinks it is a valid list) (compare `targets=??hosts??,` above)*

## Create Test File on Hosts

Create a text-file with some content on the hosts of a hostgroup.

[View Playbook](CreateFile.yml)

Requires:

- inventory-file
- vault password

Inputs:

- path - Full file-path to create the file on the target system
- content - Text to write into the file
- targets - Hostgroup-name of the hosts to create the file on

```XML
<task>
    <name>Create Testfile on Hostgroup</name>
    <command>ansible-playbook /home/ansible/ansible/playbooks/CreateFile.yml -i /home/ansible/ansible/inventory/hosts --vault-password-file /home/ansible/ansible/vars/vault.pass --extra-vars "path=??path?? content=??content?? targets=??hosts??"</command>
    <form>
        <input>
            <type>text</type>
            <label>Filepath</label>
            <var>path</var>
        </input>
        <input>
            <type>text</type>
            <label>Content</label>
            <var>content</var>
        </input>
        <input>
            <type>dropdown</type>
            <options>ubuntu;debian;lxc;hypervisor;mgmt;priv;prod;internal;external</options>
            <label>Hostgroup</label>
            <var>hosts</var>
        </input>
    </form>
</task>
```

## Configure New System

[View Playbook](ConfigureNewSystem.yml)

Requires:

- vault password
- sshd_config-file (`../vars/default-configs/sshd_config`) to copy to the server
- motd- (`../vars/default-configs/motd`) and banner-file (`../vars/default-configs/loginbanner`) to copy to th server
- Inside the Playbook
  - Path to Ansible-user's public ssh-key
  - List of management-users and their public ssh-keys
  - List of software to install
  - List of software to uninstall
  - List of services to disable

Inputs:

- targets - hosts to run the Playbook on (comma-separated without spaces OR single host)
- port - port to use for ssh connection
- user - username to login via ssh
- passwd - password to use for ssh login
- rootpasswd - password of the root account in case sudo is not installed and root cannot login via ssh (if user=root or sudo is installed and user can use it enter `0`)
- nomgmtkeys - set to `on` if no ssh-keys should be added to the defined management accounts

```XML
<task>
    <name>Neues System konfigurieren</name>
    <command>ansible-playbook /home/ansible/ansible/playbooks/ConfigureNewSystem.yml -i ??targets??, --vault-password-file /home/ansible/ansible/vars/vault.pass --extra-vars "port=??port?? user=??user?? passwd=??passwd?? rootpasswd=??rootpasswd?? nomgmtkeys=??nomgmtkeys??"</command>
    <form>
        <input>
            <type>text</type>
            <label>Host-IP (comma-separated without spaces OR single host)</label>
            <var>targets</var>
        </input>
        <input>
            <type>text</type>
            <label>SSH Port</label>
            <var>port</var>
        </input>
        <input>
            <type>text</type>
            <label>SSH Login</label>
            <var>user</var>
        </input>
        <input>
            <type>password</type>
            <label>SSH Password</label>
            <var>passwd</var>
        </input>
        <input>
            <type>password</type>
            <label>Root Pasword (if SSH-User has no sudo-permission - otherwise 0)</label>
            <var>rootpasswd</var>
        </input>
        <input>
            <type>checkbox</type>
            <label>Don't add SSH-key(s) for management-users</label>
            <var>nomgmtkeys</var>
        </input>
    </form>
</task>
```
