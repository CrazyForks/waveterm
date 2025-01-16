package fileshare

import (
	"context"
	"fmt"
	"log"

	"github.com/wavetermdev/waveterm/pkg/remote/awsconn"
	"github.com/wavetermdev/waveterm/pkg/remote/connparse"
	"github.com/wavetermdev/waveterm/pkg/remote/fileshare/fstype"
	"github.com/wavetermdev/waveterm/pkg/remote/fileshare/s3fs"
	"github.com/wavetermdev/waveterm/pkg/remote/fileshare/wavefs"
	"github.com/wavetermdev/waveterm/pkg/remote/fileshare/wshfs"
	"github.com/wavetermdev/waveterm/pkg/wshrpc"
	"github.com/wavetermdev/waveterm/pkg/wshutil"
)

// CreateFileShareClient creates a fileshare client based on the connection string
// Returns the client and the parsed connection
func CreateFileShareClient(ctx context.Context, connection string) (fstype.FileShareClient, *connparse.Connection) {
	log.Printf("CreateFileShareClient: connection=%s", connection)
	conn, err := connparse.ParseURI(connection)
	if err != nil {
		log.Printf("error parsing connection: %v", err)
		return nil, nil
	}
	if conn.Host == connparse.ConnHostCurrent {
		handler := wshutil.GetRpcResponseHandlerFromContext(ctx)
		if handler == nil {
			log.Printf("error getting rpc response handler from context")
			return nil, nil
		}
		source := handler.GetRpcContext().Conn

		// RPC context connection is empty for local connections
		if source == "" {
			source = wshrpc.LocalConnName
		}
		conn.Host = source
	}
	conntype := conn.GetType()
	if conntype == connparse.ConnectionTypeS3 {
		config, err := awsconn.GetConfig(ctx, connection)
		if err != nil {
			log.Printf("error getting aws config: %v", err)
			return nil, nil
		}
		return s3fs.NewS3Client(config), conn
	} else if conntype == connparse.ConnectionTypeWave {
		return wavefs.NewWaveClient(), conn
	} else {
		return wshfs.NewWshClient(), conn
	}
}

func Read(ctx context.Context, data wshrpc.FileData) (*wshrpc.FileData, error) {
	client, conn := CreateFileShareClient(ctx, data.Info.Path)
	if conn == nil || client == nil {
		return nil, fmt.Errorf("error creating fileshare client, could not parse connection %s", data.Info.Path)
	}
	return client.Read(ctx, conn, data)
}

func ReadStream(ctx context.Context, data wshrpc.FileData) chan wshrpc.RespOrErrorUnion[wshrpc.FileData] {
	client, conn := CreateFileShareClient(ctx, data.Info.Path)
	if conn == nil || client == nil {
		rtn := make(chan wshrpc.RespOrErrorUnion[wshrpc.FileData], 1)
		rtn <- wshrpc.RespOrErrorUnion[wshrpc.FileData]{Error: fmt.Errorf("error creating fileshare client, could not parse connection %s", data.Info.Path)}
		close(rtn)
		return rtn
	}
	return client.ReadStream(ctx, conn, data)
}

func ListEntries(ctx context.Context, path string, opts *wshrpc.FileListOpts) ([]*wshrpc.FileInfo, error) {
	log.Printf("ListEntries: path=%s", path)
	client, conn := CreateFileShareClient(ctx, path)
	if conn == nil || client == nil {
		return nil, fmt.Errorf("error creating fileshare client, could not parse connection %s", path)
	}
	return client.ListEntries(ctx, conn, opts)
}

func ListEntriesStream(ctx context.Context, path string, opts *wshrpc.FileListOpts) chan wshrpc.RespOrErrorUnion[wshrpc.CommandRemoteListEntriesRtnData] {
	log.Printf("ListEntriesStream: path=%s", path)
	client, conn := CreateFileShareClient(ctx, path)
	if conn == nil || client == nil {
		rtn := make(chan wshrpc.RespOrErrorUnion[wshrpc.CommandRemoteListEntriesRtnData], 1)
		rtn <- wshrpc.RespOrErrorUnion[wshrpc.CommandRemoteListEntriesRtnData]{Error: fmt.Errorf("error creating fileshare client, could not parse connection %s", path)}
		close(rtn)
		return rtn
	}
	return client.ListEntriesStream(ctx, conn, opts)
}

func Stat(ctx context.Context, path string) (*wshrpc.FileInfo, error) {
	log.Printf("Stat: path=%s", path)
	client, conn := CreateFileShareClient(ctx, path)
	if conn == nil || client == nil {
		return nil, fmt.Errorf("error creating fileshare client, could not parse connection %s", path)
	}
	return client.Stat(ctx, conn)
}

func PutFile(ctx context.Context, data wshrpc.FileData) error {
	log.Printf("PutFile: path=%s", data.Info.Path)
	client, conn := CreateFileShareClient(ctx, data.Info.Path)
	if conn == nil || client == nil {
		return fmt.Errorf("error creating fileshare client, could not parse connection %s", data.Info.Path)
	}
	return client.PutFile(ctx, conn, wshrpc.FileData{
		Data64: data.Data64,
		Opts:   data.Opts,
		At:     data.At,
	})
}

func Mkdir(ctx context.Context, path string) error {
	log.Printf("Mkdir: path=%s", path)
	client, conn := CreateFileShareClient(ctx, path)
	if conn == nil || client == nil {
		return fmt.Errorf("error creating fileshare client, could not parse connection %s", path)
	}
	return client.Mkdir(ctx, conn)
}

// TODO: Implement move across different fileshare types
func Move(ctx context.Context, srcPath, destPath string, recursive bool) error {
	log.Printf("Move: src=%s, dest=%s", srcPath, destPath)
	srcClient, srcConn := CreateFileShareClient(ctx, srcPath)
	if srcConn == nil || srcClient == nil {
		return fmt.Errorf("error creating fileshare client, could not parse connection %s or %s", srcPath, destPath)
	}
	destConn, err := connparse.ParseURI(destPath)
	if err != nil {
		return fmt.Errorf("error parsing destination connection %s: %v", destPath, err)
	}
	return srcClient.Move(ctx, srcConn, destConn, recursive)
}

// TODO: Implement copy across different fileshare types
func Copy(ctx context.Context, srcPath, destPath string, recursive bool) error {
	log.Printf("Copy: src=%s, dest=%s", srcPath, destPath)
	return nil
}

func Delete(ctx context.Context, path string) error {
	log.Printf("Delete: path=%s", path)
	client, conn := CreateFileShareClient(ctx, path)
	if conn == nil || client == nil {
		return fmt.Errorf("error creating fileshare client, could not parse connection %s", path)
	}
	return client.Delete(ctx, conn)
}
