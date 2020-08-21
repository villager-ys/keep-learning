### 一　ansible主要配置文件
1. ansible配置文件：</br>
/etc/ansible</br>
├── ansible.cfg</br>
├── hosts</br>
└── playbooks</br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; └── test_ping.yml

2. ansible配置文件读取顺序：

在一些特殊场景下,用户还是需要自行修改这些配置文件用户可以修改一下配置文件来修改设置,他们的被读取的顺序如下：

```
ANSIBLE_CONFIG (一个环境变量)
ansible.cfg (位于当前目录中)
.ansible.cfg (位于家目录中)
/etc/ansible/ansible.cfg
```

二　ansible.cfg配置内容解析
```
1）inventory 
该参数表示资源清单inventory文件的位置，资源清单就是一些Ansible需要连接管理的主机列表 
inventory = /root/ansible/hosts
2）library 
Ansible的操作动作，无论是本地或远程，都使用一小段代码来执行，这小段代码称为模块，这个library参数就是指向存放Ansible模块的目录 
library = /usr/share/ansible
3）forks 
设置默认情况下Ansible最多能有多少个进程同时工作，默认设置最多5个进程并行处理。具体需要设置多少个，可以根据控制主机的性能和被管理节点的数量来确定。 
forks = 5
4）default_sudo_user 
这是设置默认执行命令的用户，也可以在playbook中重新设置这个参数 
default_sudo_user = root
5）remote_port 
这是指定连接被关节点的管理端口，默认是22，除非设置了特殊的SSH端口，不然这个参数一般是不需要修改的 
remote_port = 22
6）host_key_checking 
这是设置是否检查SSH主机的密钥。可以设置为True或False 
host_key_checking = False
7）timeout 
这是设置SSH连接的超时间隔，单位是秒。 
timeout = 20
8）log_path 
Ansible系统默认是不记录日志的，如果想把Ansible系统的输出记录到人i治稳健中，需要设置log_path来指定一个存储Ansible日志的文件 
log_path = /var/log/ansible.log
另外需要注意，执行Ansible的用户需要有写入日志的权限，模块将会调用被管节点的syslog来记录，口令是不会出现的日志中的
9）private_key_file
在使用ssh公钥私钥登录系统时候，使用的密钥路径。
private_key_file=/path/to/file.pem
```
 [详细配置信息，参开官网！](https://docs.ansible.com/ansible/latest/installation_guide/intro_configuration.html)
 
###  三　hosts文件配置方式
```
[work]
192.168.10.150 absible_ssh_user=root ansible_ssh_pass='123456'
test1 ansible_ssh_host=192.168.10.99 ansible_ssh_host=22 ansible_ssh_user=root ansible_ssh_private_key_file=/home/ssh_keys/id_rsa
# 别名+ssh用户+ssh秘钥
```
执行ansible all --list-hosts查看所有主机资产
```
yuanshuai-ThinkPad-T480s% ansible all --list-hosts  
  hosts (2):
    192.168.10.150
    192.168.10.99
```
### 四　ansible命令中ad-hoc模式使用的参数介绍
1. -v, --verbose：输出更详细的执行过程信息，-vvv可得到所有执行过程信息。
2. -i PATH, --inventory=PATH：指定inventory信息，默认/etc/ansible/hosts。
3. -f NUM, --forks=NUM：并发线程数，默认5个线程。
4. --private-key=PRIVATE_KEY_FILE：指定密钥文件。
5. -m NAME, --module-name=NAME：指定执行使用的模块.
6. -M DIRECTORY, --module-path=DIRECTORY：指定模块存放路径，默认/usr/share/ansible，也可以通过ANSIBLE_LIBRARY设定默认路径。
7. -a 'ARGUMENTS', --args='ARGUMENTS'：模块参数。
8. -k, --ask-pass SSH：认证密码。
9. -K, --ask-sudo-pass sudo：用户的密码（—sudo时使用）。
10. -o, --one-line：标准输出至一行。
11. -s, --sudo：相当于Linux系统下的sudo命令。
12. -t DIRECTORY, --tree=DIRECTORY：输出信息至DIRECTORY目录下，结果文件以远程主机名命名。
13. -T SECONDS, --timeout=SECONDS：指定连接远程主机的最大超时，单位是：秒。
14. -B NUM, --background=NUM：后台执行命令，超NUM秒后kill正在执行的任务。
15. -P NUM, --poll=NUM：定期返回后台任务进度。
16. -u USERNAME, --user=USERNAME：指定远程主机以USERNAME运行命令。
17. -U SUDO_USERNAME, --sudo-user=SUDO_USERNAM：E使用sudo，相当于Linux下的sudo命令。
18. -c CONNECTION, --connection=CONNECTION：指定连接方式，可用选项paramiko (SSH), ssh, local。Local方式常用于crontab 和 kickstarts。
19. -l SUBSET, --limit=SUBSET：指定运行主机。
20. -l ~REGEX, --limit=~REGEX：指定运行主机（正则）。
21. --list-hosts：列出符合条件的主机列表，不执行任何其他命令

### 五　-m常用模块

模块名 | 作用 | 案例
---|---|---
command | 默认模块，执行命令 | ansible test -a "ls /tmp"
shell | 执行shell命令，可以使用$,\|等 | ansible test -m shell -a "echo $HOSTNAME"
copy | 文件传输 | ansibel test -m copy -a "src=/etc/hosts dest=/tmp/hosts"
yum | yum包管理 | ansibel test -m yum -a "name=nginx state=present"
user | 用户 | ansible test -m user -a "name=yuanshuai password=123456"
git | git模块 | ansible test -m git -a "repo=https://github.com/xx dest=/opt/ops version=HEAD" 
service | 服务管理 | ansible test -m service -a "name=nginx state=started"



### 六 playbook
##### playbook的组成：
&nbsp;&nbsp;&nbsp;&nbsp;play:定义的是主机的角色

&nbsp;&nbsp;&nbsp;&nbsp;task:定义具体执行的任务

&nbsp;&nbsp;&nbsp;&nbsp;playbook:由一个或多个play组成，一个play可以包含多个task
```
---
- hosts: work
  gather_facts: False
  tasks:

    - name: test ping
      ping:
    - name: shell commond
      shell: echo "hello world"
      register: result
    - name: show debug info
      debug: var=result.stdout verbosity=0

```
##### playbook常用命令参数：
执行方式：ansible-playbook playbook.yml [options]
```
＃ ssh 连接的用户名
-u REMOTE_USER, --user=REMOTE_USER  

＃ssh登录认证密码
 -k, --ask-pass    

＃sudo 到root用户，相当于Linux系统下的sudo命令
 -s, --sudo           

＃sudo 到对应的用户
 -U SUDO_USER, --sudo-user=SUDO_USER    

＃用户的密码（—sudo时使用）
 -K, --ask-sudo-pass     

＃ ssh 连接超时，默认 10 秒
 -T TIMEOUT, --timeout=TIMEOUT 

＃ 指定该参数后，执行 playbook 文件不会真正去执行，而是模拟执行一遍，然后输出本次执行会对远程主机造成的修改
 -C, --check      

＃ 设置额外的变量如：key=value 形式 或者 YAML or JSON，以空格分隔变量，或用多个-e
 -e EXTRA_VARS, --extra-vars=EXTRA_VARS    

＃ 进程并发处理，默认 5
 -f FORKS, --forks=FORKS    

＃ 指定 hosts 文件路径，默认 default=/etc/ansible/hosts
 -i INVENTORY, --inventory-file=INVENTORY   

＃ 指定一个 pattern，对-hosts:匹配到的主机再过滤一次
 -l SUBSET, --limit=SUBSET    

＃ 只打印有哪些主机会执行这个 playbook文件，不是实际执行该 playbook
 --list-hosts  

＃ 列出该 playbook 中会被执行的 task
 --list-tasks   

＃ 私钥路径
 --private-key=PRIVATE_KEY_FILE   

＃ 同一时间只执行一个 task，每个 task执行前都会提示确认一遍
 --step    

＃ 只检测 playbook 文件语法是否有问题，不会执行该 playbook 
 --syntax-check  

＃当 play 和 task 的 tag为该参数指定的值时才执行，多个 tag 以逗号分隔
 -t TAGS, --tags=TAGS   

＃ 当 play 和 task 的 tag不匹配该参数指定的值时，才执行
 --skip-tags=SKIP_TAGS   

＃输出更详细的执行过程信息，-vvv可得到所有执行过程信息。
 -v, --verbose   
```

#### playbook里定义变量
1. playbook的yaml中定义变量赋值
2. --extra-vars执行命令是赋值
3. 在文件中定义变量(资产清单文件/etc/ansile/hosts)
4. 注册变量(register)
```
---
- hosts: work
  remote_user: root
  vars:
    touch_file: a.file
  tasks:
    - name: get date
      command: date
      register: date_output
    - name: echo date_output
      shell : "echo {{date_output.stdout}}>/tmp/{{touch_file}}"

```
#### playbook中的语法
when条件语句(if语句)，它可以用来控制playbooks中任务的执行流程。类似于程序中的if条件语句一样，使得ansible可以更好的按照运维人员的意愿来对远程节点执行特定的操作:
简单事例
```
    tasks:
     - name: "shutdown Debian flavored systems"
        command: /sbin/shutdown -t now
        when: ansible_os_family == "Debian
```

循环语句：
1. 标准循环

添加多个用户
```
- name: add several users
  user: name={{ item }} state=present groups=wheel
  with_items:
     - testuser1
     - testuser2
```
添加多个用户，并将用户加入不同的组内
```
- name: add several users
  user: name={{ item.name }} state=present groups={{ item.groups }}
  with_items:
    - { name: 'testuser1', groups: 'wheel' }
    - { name: 'testuser2', groups: 'root' }
```
2. 嵌套循环

分别给用户授予3个数据库的所有权限
```
- name: give users access to multiple databases
  mysql_user: name={{ item[0] }} priv={{ item[1] }}.*:ALL append_privs=yes password=foo
  with_nested:
    - [ 'alice', 'bob' ]
    - [ 'clientdb', 'employeedb', 'providerdb' ]
```

3. 遍历字典

输出用户的姓名和电话
```
tasks:
  - name: Print phone records
    debug: msg="User {{ item.key }} is {{ item.value.name }} ({{ item.value.telephone }})"
    with_dict: {'alice':{'name':'Alice Appleworth', 'telephone':'123-456-789'},'bob':{'name':'Bob Bananarama', 'telephone':'987-654-3210'} }
```
4. 并行遍历列表

如果列表数目不匹配，用None补全
```
tasks:
  - debug: "msg={{ item.0 }} and {{ item.1 }}"
    with_together:
    - [ 'a', 'b', 'c', 'd','e' ]
    - [ 1, 2, 3, 4 ]

```
5. 遍历列表和索引

item.0 为索引，item.1为值
```
 - name: indexed loop demo
    debug: "msg='at array position {{ item.0 }} there is a value {{ item.1 }}'"
    with_indexed_items: [1,2,3,4]
```
6. 遍历文件列表的内容

```
---
- hosts: all
  tasks:
       - debug: "msg={{ item }}"
      with_file:
        - first_example_file
        - second_example_file
```
7. 遍历目录文件

with_fileglob匹配单个目录中的所有文件，非递归匹配模式。当在role中使用with_fileglob的相对路径时，Ansible解析相对于roles/<rolename>/files目录的路径。

```
- hosts: all
  tasks:
    - file: dest=/etc/fooapp state=directory
    - copy: src={{ item }} dest=/etc/fooapp/ owner=root mode=600
      with_fileglob:
        - /playbooks/files/fooapp/*
```

8. 遍历ini文件

lookup.ini

[section1]

value1=section1/value1

value2=section1/value2

[section2]

value1=section2/value1

value2=section2/value2
```
- debug: msg="{{ item }}"
  with_ini: value[1-2] section=section1 file=lookup.ini re=true

```
获取section1 里的value1和value2的值
9. 重试循环 until

"重试次数retries" 的默认值为3，"delay"为5。
```
- action: shell /usr/bin/foo
  register: result
  until: result.stdout.find("all systems go") != -1
  retries: 5
  delay: 10
```
10. 查找第一个匹配文件

依次寻找列表中的文件，找到就返回。如果列表中的文件都找不到，任务会报错。
```
  tasks:
  - debug: "msg={{ item }}"
    with_first_found:
     - "/tmp/a"
     - "/tmp/b"
     - "/tmp/default.conf"
```
11. 随机选择with_random_choice

随机选择列表中得一个值
```
- debug: msg={{ item }}
  with_random_choice:
     - "go through the door"
     - "drink from the goblet"
     - "press the red button"
     - "do nothing"
```

12. 循环子元素

定义好变量
```
#varfile
---
users:
  - name: alice
    authorized:
      - /tmp/alice/onekey.pub
      - /tmp/alice/twokey.pub
    mysql:
        password: mysql-password
        hosts:
          - "%"
          - "127.0.0.1"
          - "::1"
          - "localhost"
        privs:
          - "*.*:SELECT"
          - "DB1.*:ALL"
  - name: bob
    authorized:
      - /tmp/bob/id_rsa.pub
    mysql:
        password: other-mysql-password
        hosts:
          - "db1"
        privs:
          - "*.*:SELECT"
          - "DB2.*:ALL"
```
```
---
- hosts: web
  vars_files: varfile
  tasks:

  - user: name={{ item.name }} state=present generate_ssh_key=yes
    with_items: "{{ users }}"

  - authorized_key: "user={{ item.0.name }} key='{{ lookup('file', item.1) }}'"
    with_subelements:
      - "{{ users }}"
      - authorized

  - name: Setup MySQL users
    mysql_user: name={{ item.0.name }} password={{ item.0.mysql.password }} host={{ item.1 }} priv={{ item.0.mysql.privs | join('/') }}
    with_subelements:
      - "{{ users }}"
      - mysql.hosts
```
{{ lookup('file', item.1) }} 是查看item.1文件的内容

with_subelements 遍历哈希列表，然后遍历列表中的给定（嵌套）的键。

13. 在序列中循环with_sequence

with_sequence以递增的数字顺序生成项序列。 您可以指定开始，结束和可选步骤值。 参数应在key = value对中指定。'format'是一个printf风格字符串。数字值可以以十进制，十六进制（0x3f8）或八进制（0600）指定。 不支持负数。
```
---
- hosts: all

  tasks:

    # 创建组
    - group: name=evens state=present
    - group: name=odds state=present

    # 创建格式为testuser%02x 的0-32 序列的用户
    - user: name={{ item }} state=present groups=evens
      with_sequence: start=0 end=32 format=testuser%02x

    # 创建4-16之间得偶数命名的文件
    - file: dest=/var/stuff/{{ item }} state=directory
      with_sequence: start=4 end=16 stride=2

    # 简单实用序列的方法：创建4 个用户组分表是组group1 group2 group3 group4
    - group: name=group{{ item }} state=present
      with_sequence: count=4
```

#### hanlders和tags
1. handlers

在需要被监控的任务（tasks）中定义一个notify，只有当这个任务被执行时，才会触发notify对应的handlers去执行相应操作。
```
---
- hosts: control-node
  remote_user: root
  vars:
    - pkg: httpd
  tasks:
    - name: "install httpd package."
      yum: name={{ pkg }}  state=installed
    - name: "copy httpd configure file to remote host."
      copy: src=/root/conf/httpd.conf dest=/etc/httpd/conf/httpd.conf
      notify: restart httpd
    - name: "boot httpd service."
      service: name=httpd state=started
  handlers:
    - name: restart httpd
      service: name=httpd state=restarted
```
在使用handlers的过程中，有以下几点需要注意：

handlers只有在其所在的任务被执行时，都会被运行；

handlers只会在Play的末尾运行一次；如果想在一个Playbook的中间运行handlers，则需要使用meta模块来实现，例如：- meta: flush_handlers。

如果一个Play在运行到调用handlers的语句之前失败了，那么这个handlers将不会被执行。我们可以使用mega模块的--force-handlers选项来强制执行handlers，即使在handlers所在Play中途运行失败也能执行。

2. tags

tags用于让用户选择运行playbook中的部分代码。ansible具有幂等性，因此会自动跳过没有变化的部分，即便如此，有些代码为测试其确实没有发生变化的时间依然会非常地长。此时，如果确信其没有变化，就可以通过tags跳过此些代码片断。

ansible的标签（Tags）功能可以给角色（Roles）、文件、单独的任务，甚至整个Playbook打上标签，然后利用这些标签来指定要运行Playbook中的个别任务，或不执行指定的任务。
```
---
- hosts: control-node
  remote_user: root
  vars:
    - pkg: httpd
  tasks:
    - name: "install httpd package."
      yum: name={{ pkg }}  state=installed
    - name: "copy httpd configure file to remote host."
      copy: src=/root/conf/httpd.conf dest=/etc/httpd/conf/httpd.conf
      notify: restart httpd
    - name: "start httpd service."
      tags:
        - start_httpd     # 给“start httpd service”这个任务打个标签
      service: name=httpd state=started
  handlers:
    - name: restart httpd
      service: name=httpd state=restarted
```
使用：
```
[root@LOCALHOST ~]# ansible-playbook yaml/httpd.yaml --tags start_httpd

PLAY [control-node] ********************************************************************************************************

TASK [Gathering Facts] *****************************************************************************************************
ok: [openstack-control1]
ok: [openstack-control2]

TASK [start httpd service.] ************************************************************************************************
ok: [openstack-control1]
ok: [openstack-control2]

PLAY RECAP *****************************************************************************************************************
openstack-control1         : ok=2    changed=0    unreachable=0    failed=0   
openstack-control2         : ok=2    changed=0    unreachable=0    failed=0 
```
#### 异常处理
1. ignore_errors
 
在有些情况下，一些必须运行的命令或脚本会报一些错误，而这些错误并不一定真的说明有问题，但是经常会给接下来要运行的任务造成困扰，甚至直接导致playbook运行中断。

这时候，我们可以在相关任务中添加ignore_errors: true来屏蔽当前任务的报错信息。ansible也将视该任务运行成功，不再报错，这样就不会对接下来要运行的任务造成额外困扰。但是要注意的是，我们不应过度依赖ignore_errors，因为它会隐藏所有的报错信息，而应该把精力集中在寻找报错的原因上面，这样才能从根本上解决问题。
```
---
- hosts: load-node
  remote_user: root
  vars:
    - pkg: httpd
  tasks:
    - name: "install httpd package."
      yum: name={{ pkg }}  state=installed
    - name: "copy httpd configure file to remote host."
      copy: src=/root/config/httpd.conf dest=/etc/httpd/conf/httpd.conf
      notify: restart httpd
      ignore_errors: true         # 忽略错误
    - name: "start httpd service."
      tags:
        - start_httpd
      service: name=httpd state=started
  handlers:
    - name: restart httpd
      service: name=httpd state=restarted
```
2. failed_when

当满足一定的条件时，主动抛出错误
```
- hosts: DH-TEST
  remote_user: root
  gather_facts: false
  tasks:
  - name: get process
    shell: ps aux | wc -l 
    register: process_count
    failed_when: process_count > 3
  - name: touch a file
    file: path=/tmp/test3.txt state=touch owner=root mode=0700
```
failed_when: process_count > 3当进程数大于3时主动抛出错误，后续任务就不会执行了。如果不满足条件，则不会抛出错误。

3. changed_when

主机状态发生改变，不会在报黄色
```
---
- hosts: DH-TEST
  remote_user: root
  gather_facts: false
  tasks:
  - name: touch a file
    file: path=/tmp/changed_testi2 state=touch
    changed_when: false       # 关闭状态改变提示
```
### 七 Roles
Roles是一种利用在大型Playbook中的剧本配置模式，它有着自己特定的结构。用于层次性、结构化地组织playbook。roles能够根据层次型结构自动装载变量文件、tasks以及handlers等。要使用roles只需要在playbook中使用include指令即可。简单来讲，roles就是通过分别将变量、文件、任务、模板及处理器放置于单独的目录中，并可以便捷地include它们的一种机制。角色一般用于基于主机构建服务的场景中，但也可以是用于构建守护进程等场景中。

一个roles的案例如下所示：
```
site.yml            # 主入口文件
webservers.yml      # webserver类型服务所用的剧本
dbservers.yml       # 数据库类型的服务所用的剧本
files/              # 存放通用的将要被上传的文件
templates/          # 存放通用的模板文件
roles/              # roles目录名称是固定的
   common/          # 此目录下的各个组件是所有角色共用的
     tasks/         # 存放通用的任务文件
     handlers/      # 存放通用的处理器文件
     vars/          # 存放通用的变量文件 
     meta/          # 存放通用的角色依赖文件
   webservers/      # 存放webserver类型的服务的各个组件  
     files/         # 存放webserver角色需要的上传文件
     templates/     # 存放webserver角色需要的模板文件
     tasks/         # 存放webserver角色任务文件
     handlers/      # 存放webserver角色处理器
     vars/          # 存放webserver角色变量
     meta/          
```
而在playbook中，可以这样使用roles：
```
---
- hosts: webservers
  roles:
     - common
     - webservers
```
也可以向roles传递参数，例如：
```
---
- hosts: webservers
  roles:
    - common
    - { role: foo_app_instance, dir: '/opt/a',  port: 5000 }
    - { role: foo_app_instance, dir: '/opt/b',  port: 5001 }
```
7.1 创建role的步骤

(1) 创建以roles命名的目录；

(2) 在roles目录中分别创建以各角色名称命名的目录，如webservers等；

(3) 在每个角色命名的目录中分别创建files、handlers、meta、tasks、templates和vars目录；用不到的目录可以创建为空目录，也可以不创建；

(4) 在roles目录的同级目录下创建一个yaml文件，如：site.yml ，在此文件中调用各角色；

7.2 role内各目录中可用的文件

tasks目录：至少应该包含一个名为main.yml的文件，其定义了此角色的任务列表；此文件可以使用include包含其它的位于此目录中的task文件；

files目录：存放由copy或script等模块调用的文件；

templates目录：template模块会自动在此目录中寻找Jinja2模板文件；

handlers目录：此目录中应当包含一个main.yml文件，用于定义此角色用到的各handler；在handler中使用include包含的其它的handler文件也应该位于此目录中；

vars目录：应当包含一个main.yml文件，用于定义此角色用到的变量；

meta目录：应当包含一个main.yml文件，用于定义此角色的特殊设定及其依赖关系；ansible 1.3及其以后的版本才支持；

default目录：为当前角色设定默认变量时使用此目录；应当包含一个main.yml文件；

==简单案例==：

先看一下这个roles的目录结构：
```
lamp_simple
├── group_vars
│   ├── all
│   └── dbservers
├── hosts
├── LICENSE.md
├── README.md
├── roles
│   ├── common
│   │   ├── handlers
│   │   │   └── main.yml
│   │   ├── tasks
│   │   │   └── main.yml
│   │   └── templates
│   │       └── ntp.conf.j2
│   ├── db
│   │   ├── handlers
│   │   │   └── main.yml
│   │   ├── tasks
│   │   │   └── main.yml
│   │   └── templates
│   │       └── my.cnf.j2
│   └── web
│       ├── handlers
│       │   └── main.yml
│       ├── tasks
│       │   ├── copy_code.yml
│       │   ├── install_httpd.yml
│       │   └── main.yml
│       └── templates
│           └── index.php.j2
└── site.yml

14 directories, 17 files
```

查看各个playbook的内容：

查看主机清单文件
```
cat lamp_simple/hosts
[webservers]
web3

[dbservers]
web2
```

查看主入口文件
```
# cat lamp_simeple/site.yml
---
# This playbook deploys the whole application stack in this site.

- name: apply common configuration to all nodes
  hosts: all
  remote_user: root

  roles:
    - common

- name: configure and deploy the webservers and application code
  hosts: webservers
  remote_user: root

  roles:
    - web

- name: deploy MySQL and configure the databases
  hosts: dbservers
  remote_user: root

  roles:
    - db
```
查看变量文件：
```
# cat lamp_simple/group_vars/all
---
# Variables listed here are applicable to all host groups

httpd_port: 80
ntpserver: 192.168.1.2
repository: https://github.com/bennojoy/mywebapp.git
```
```
# cat lamp_simple/group_vars/dbservers
---
# The variables file used by the playbooks in the dbservers group.
# These don't have to be explicitly imported by vars_files: they are autopopulated.

mysqlservice: mysqld
mysql_port: 3306
dbuser: foouser
dbname: foodb
upassword: abc
```
查看通用hanlder文件：
```
# cat lamp_simple/roles/common/handlers/main.yml
---
# Handler to handle common notifications. Handlers are called by other plays.
# See http://docs.ansible.com/playbooks_intro.html for more information about handlers.

- name: restart ntp
  service: name=ntpd state=restarted

- name: restart iptables
  service: name=iptables state=restarted
```
查看通用tasks文件：
```
# cat lamp_simple/roles/common/tasks/main.yml
---
# This playbook contains common plays that will be run on all nodes.

- name: Install ntp
  yum: name=ntp state=present
  tags: ntp

- name: Configure ntp file
  template: src=ntp.conf.j2 dest=/etc/ntp.conf
  tags: ntp
  notify: restart ntp

- name: Start the ntp service
  service: name=ntpd state=started enabled=yes
  tags: ntp

- name: test to see if selinux is running
  command: getenforce
  register: sestatus
  changed_when: false
```
查看通用模板文件：
```
driftfile /var/lib/ntp/drift

restrict 127.0.0.1 
restrict -6 ::1

server {{ ntpserver }}

includefile /etc/ntp/crypto/pw

keys /etc/ntp/keys
```
查看db角色的handlers文件：
```
---
# Handler to handle DB tier notifications

- name: restart mysql
  service: name=mysqld state=restarted

- name: restart iptables
  service: name=iptables state=restarted[root@LOCALHOST ansible-examples-master]# cat lamp_simple/roles/db/handlers/main.yml 
---
# Handler to handle DB tier notifications

- name: restart mysql
  service: name=mysqld state=restarted

- name: restart iptables
  service: name=iptables state=restarted
```
查看db角色的tasks文件：
```
---
# This playbook will install mysql and create db user and give permissions.

- name: Install Mysql package
  yum: name={{ item }} state=installed
  with_items:
   - mysql-server
   - MySQL-python
   - libselinux-python
   - libsemanage-python

- name: Configure SELinux to start mysql on any port
  seboolean: name=mysql_connect_any state=true persistent=yes
  when: sestatus.rc != 0

- name: Create Mysql configuration file
  template: src=my.cnf.j2 dest=/etc/my.cnf
  notify:
  - restart mysql

- name: Start Mysql Service
  service: name=mysqld state=started enabled=yes

- name: insert iptables rule
  lineinfile: dest=/etc/sysconfig/iptables state=present regexp="{{ mysql_port }}"
              insertafter="^:OUTPUT " line="-A INPUT -p tcp  --dport {{ mysql_port }} -j  ACCEPT"
  notify: restart iptables

- name: Create Application Database
  mysql_db: name={{ dbname }} state=present

- name: Create Application DB User
  mysql_user: name={{ dbuser }} password={{ upassword }} priv=*.*:ALL host='%' state=present
```
查看db角色的模板文件：
```
[mysqld]
datadir=/var/lib/mysql
socket=/var/lib/mysql/mysql.sock
user=mysql
# Disabling symbolic-links is recommended to prevent assorted security risks
symbolic-links=0
port={{ mysql_port }}

[mysqld_safe]
log-error=/var/log/mysqld.log
pid-file=/var/run/mysqld/mysqld.pid
```
查看web角色的handlers文件：
```
---
# Handler for the webtier: handlers are called by other plays.
# See http://docs.ansible.com/playbooks_intro.html for more information about handlers.

- name: restart iptables
  service: name=iptables state=restarted
```
查看web角色的tasks文件：
```
---
- include: install_httpd.yml
- include: copy_code.yml
```
```
---
# These tasks install http and the php modules.

- name: Install http and php etc
  yum: name={{ item }} state=present
  with_items:
   - httpd
   - php
   - php-mysql
   - git
   - libsemanage-python
   - libselinux-python

- name: insert iptables rule for httpd
  lineinfile: dest=/etc/sysconfig/iptables create=yes state=present regexp="{{ httpd_port }}" insertafter="^:OUTPUT "
              line="-A INPUT -p tcp  --dport {{ httpd_port }} -j  ACCEPT"
  notify: restart iptables

- name: http service state
  service: name=httpd state=started enabled=yes

- name: Configure SELinux to allow httpd to connect to remote database
  seboolean: name=httpd_can_network_connect_db state=true persistent=yes
  when: sestatus.rc != 0`
```
```
---
# These tasks are responsible for copying the latest dev/production code from
# the version control system.

- name: Copy the code from repository
  git: repo={{ repository }} dest=/var/www/html/

- name: Creates the index.php file
  template: src=index.php.j2 dest=/var/www/html/index.php
```
查看web角色的模板文件：
```
<html>
 <head>
  <title>Ansible Application</title>
 </head>
 <body>
 </br>
  <a href=http://{{ ansible_default_ipv4.address }}/index.html>Homepage</a>
 </br>
<?php 
 Print "Hello, World! I am a web server configured using Ansible and I am : ";
 echo exec('hostname');
 Print  "</BR>";
echo  "List of Databases: </BR>";
        {% for host in groups['dbservers'] %}
                $link = mysqli_connect('{{ hostvars[host].ansible_default_ipv4.address }}', '{{ hostvars[host].dbuser }}', '{{ hostvars[host].upassword }}') or die(mysqli_connect_error($link));
        {% endfor %}
        $res = mysqli_query($link, "SHOW DATABASES;");
        while ($row = mysqli_fetch_assoc($res)) {
                echo $row['Database'] . "\n";
        }
?>
</body>
</html>
```
执行这个roles
```
ansile-playbook -i lamp_simple/hosts lamp_simple/site.yml
```

### 八、Ansible Galaxy
Ansible Galaxy是Ansible官方Roles分享平台（galaxy.ansible.com），在Galaxy平台上所有人可以免费上传或下载Roles，在这里好的技巧、思想、架构得以积累和传播。

相关命令：
```
搜索nginx相关的roles:
ansible-galaxy search nginx

安装一个搜索到的角色：
ansible-galaxy install 1davidmichael.ansible-role-nginx
```