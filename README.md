# Ansible Web Controller (AWC)

AWC is a web application written in GO that allows you to execute preconfigured tasks as a user of your Linux system. The idea is to run Ansible Playbooks and view the results remotely. Arguments can be supplied using HTML-Forms. The forms to fill in and commands to execut will be generated dynamically based on the `commands.xml` file. Beside Playbooks technically any Linux command can be configured as a task.

## Annotations

This application is NOT designed for production use! You may use it for testing or in home lab environments.  
DO NOT run this application as root! Once configured the commands can be executed with any parameter. You have to make sure that the parameter will not endanger your system. I recommend using a non root user that can only connect to the Ansible clients via SSH.  
Currently there is no authentication included in this application. I recommend blocking the application's port via a firewall and only allow access to the application from localhost or a reverse proxy with authentication (see Installation).

## Setup

### Environment

### Installation

### Configuration

## License

This application is published undes GNU GNU GENERAL PUBLIC LICENSE Version 3 as refered in the LICENSE-File.

Copyright (C) 2022 0Raptor

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not, see [www.gnu.org/licenses/](https://www.gnu.org/licenses/).
