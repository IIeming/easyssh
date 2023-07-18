package authorized

import (
	"easyssh/config"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"golang.org/x/crypto/ssh"
)

var (
	password      string
	publicKeyPath string
	sshMode       ssh.AuthMethod
	publicKey     []byte
)

func Init(hosts *config.Config, host *config.Hosts) {
	// 确定ssh连接方式
	// 密钥连接
	if hosts.IsSecret {
		if hosts.PublicKey == "" {
			publicKeyPath = host.PublicKey
		} else {
			publicKeyPath = hosts.PublicKey
		}
		privateKeyBytes, err := os.ReadFile(publicKeyPath)
		if err != nil {
			log.Fatalf("读取密钥文件%s失败: %v\n", publicKeyPath, err)
		}
		// 编码密钥格式
		privateKey, err := ssh.ParsePrivateKey(privateKeyBytes)
		if err != nil {
			log.Fatal("解析私钥文件失败: ", err)
		}
		sshMode = ssh.PublicKeys(privateKey)
	} else {
		// 密码连接
		// 服务器列表密码是否都一致
		if hosts.Passwd == "" {
			password = host.Passwd
		} else {
			password = hosts.Passwd
		}
		sshMode = ssh.Password(password)
	}

	// 创建SSH配置
	config := &ssh.ClientConfig{
		User: hosts.User,
		Auth: []ssh.AuthMethod{
			sshMode,
		},
		// 连接超时时间
		Timeout: time.Second * 10,
		// 忽略主机密钥验证
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// 连接到远程服务器
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host.Host, hosts.Port), config)
	if err != nil {
		log.Fatalf("连接到 %s:%d 失败: %v ", host.Host, hosts.Port, err)
	}
	// 关闭ssh连接
	defer conn.Close()

	// 创建一个新的会话(Session)，执行命令
	session, err := conn.NewSession()
	if err != nil {
		log.Fatalf("创建会话 session 失败: %v", err)
	}
	defer session.Close()

	// 创建一个新的会话，执行命令(一个session只能执行一次run函数)
	session2, err := conn.NewSession()
	if err != nil {
		log.Fatalf("创建会话 session2 失败: %v", err)
	}
	defer session2.Close()

	// 在主服务器上生成公钥
	if hosts.Hosts[0].Name == "worker-0" {
		rsapubPath := filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa.pub")
		_, err = os.Stat(rsapubPath)
		if os.IsNotExist(err) {
			cmdString := "ssh-keygen -f " + rsapubPath + "-t rsa -P ''"
			err = session.Run(cmdString)
			if err != nil {
				log.Fatalf("生成公钥 id_rsa.pub 文件失败: %v", err)
			}
		}

		// 获取密钥内容
		publicKeyFile := filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa.pub")
		publicKey, err = os.ReadFile(publicKeyFile)
		if err != nil {
			log.Fatalf("读取公钥文件失败: %v", err)
		}
	} else {
		log.Fatalln("worker-0必须排在第一位")
	}

	// 写入公钥到免密机器
	sshPath := filepath.Join(os.Getenv("HOME"), ".ssh")
	authorizedKeysFile := filepath.Join(os.Getenv("HOME"), ".ssh", "authorized_keys")
	err = session2.Run(fmt.Sprintf("mkdir -m 700 %s;", sshPath) + fmt.Sprintf("echo '%s' >> %s", publicKey, authorizedKeysFile))
	if err != nil {
		log.Println(fmt.Sprintf("主机 %v 免密配置", host.Host), color.RedString("失败"))
	}

	log.Println(fmt.Sprintf("主机 %v 免密配置", host.Host), color.GreenString("成功"))
}
