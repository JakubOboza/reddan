package redis

// SPEC: http://redis.io/topics/protocol
// LINE_END =\r\n

import (
	"bufio"
	"bytes"
	"errors"
	"net"
	"strconv"
)

const (
	bufSize int = 4096
)

var CLRF = "\r\n"

type Client struct {
	connection net.Conn
}

// Example
// redis.Dial("localhost:6379")
func Dial(addr string) (*Client, error) {

	conn, err := net.Dial("tcp", addr)

	if err != nil {
		return nil, err
	}

	client := new(Client)
	client.connection = conn

	return client, nil

}

// Example
// defer client.Close()
func (client *Client) Close() error {
	if client.connection == nil {
		return errors.New("Can't close empty connection")
	}
	return client.connection.Close()
}

func readResponse(reader *bufio.Reader) (interface{}, error) {

	response_type, err := reader.ReadByte()

	if err != nil {
		return nil, err
	}

	switch response_type {
	case '-':
		return readError(reader)
	case '$':
		return readString(reader)
	case '+':
		return readSimpleString(reader)
	case ':':
		return readInteger(reader)
	case '*':
		return readArray(reader)
	default:
		return nil, errors.New("Parse error, could not parse the response")
	}

}

func readArray(reader *bufio.Reader) ([]interface{}, error) {
	elements_to_read, err := reader.ReadBytes('\n')

	if err != nil {
		return nil, errors.New("Error while reading response from server")
	}

	elements_to_read = elements_to_read[0 : len(elements_to_read)-2] //getting rid of trailing \r\n
	number_of_elements, err := strconv.Atoi(string(elements_to_read))

	result := make([]interface{}, number_of_elements)

	for i := 0; i < number_of_elements; i++ {
		res, _ := readResponse(reader)
		// TODO add handling of errors here, possible refactor of array response handling
		result[i] = res
	}

	return result, nil
}

func readString(reader *bufio.Reader) ([]byte, error) {
	bytes_to_read, err := reader.ReadBytes('\n')

	if err != nil {
		return nil, err
	}

	bytes_to_read = bytes_to_read[0 : len(bytes_to_read)-2] //getting rid of trailing \r\n
	num_bytes, err := strconv.Atoi(string(bytes_to_read))

	if num_bytes < 0 {
		return nil, errors.New("Key couldn't be found")
	}

	if err != nil {
		return nil, err
	}

	string_line, err := reader.ReadBytes('\n')

	return string_line[0:num_bytes], nil
}

// reades errors like -ERR not working :D\r\n
func readError(reader *bufio.Reader) ([]byte, error) {
	raw_response, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	return nil, errors.New(string(raw_response[0 : len(raw_response)-2]))
}

// parses Simple string response like +OK\r\n
func readSimpleString(reader *bufio.Reader) ([]byte, error) {
	raw_response, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	return raw_response[0 : len(raw_response)-2], nil
}

func readInteger(reader *bufio.Reader) ([]byte, error) {
	raw_response, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	return raw_response[0 : len(raw_response)-2], nil
}

// Builds a command that will be written to the redis server connection
func buildCommand(cmd string, args ...string) []byte {

	var buffer bytes.Buffer

	buffer.WriteString(cmd)

	for _, arg := range args {
		buffer.WriteString(" ")
		buffer.WriteString(strconv.Quote(arg))
	}

	buffer.WriteString(CLRF)

	return buffer.Bytes()

}

func (client *Client) runCommand(command []byte) (int, error) {
	return client.connection.Write(command)
}

func (client *Client) executeKeyCommand(command []byte) (string, error) {
	_, err := client.runCommand(command)

	if err != nil {
		return "", err
	}

	reader := bufio.NewReaderSize(client.connection, bufSize)
	result, err := readResponse(reader)

	if res, ok := result.([]byte); ok {
		return string(res), err
	} else {
		return "", errors.New("Internal reddan error, could not recognize response type")
	}

}

func (client *Client) executeBoolCommand(command []byte) (bool, error) {
	_, err := client.runCommand(command)

	if err != nil {
		return false, err
	}

	reader := bufio.NewReaderSize(client.connection, bufSize)
	result, err := readResponse(reader)

	if err != nil {
		return false, err
	}

	if res, ok := result.([]byte); ok {
		return strconv.ParseBool(string(res))
	} else {
		return false, errors.New("Internal reddan error, could not recognize response type")
	}

}

func (client *Client) executeIntCommand(command []byte) (int, error) {
	_, err := client.runCommand(command)

	if err != nil {
		return 0, err
	}

	reader := bufio.NewReaderSize(client.connection, bufSize)
	result, err := readResponse(reader)

	if err != nil {
		return 0, err
	}

	if res, ok := result.([]byte); ok {
		return strconv.Atoi(string(res))
	} else {
		return 0, errors.New("Internal reddan error, unexpected response type")
	}

}

func (client *Client) executeStringArrayCommand(command []byte) ([]string, error) {
	_, err := client.runCommand(command)

	if err != nil {
		return nil, err
	}

	reader := bufio.NewReaderSize(client.connection, bufSize)
	result, err := readResponse(reader)

	if err != nil {
		return nil, err
	}

	if res, ok := result.([]interface{}); ok {

		result_array := make([]string, len(res))

		for i, elem := range res {
			el := elem.([]byte)
			result_array[i] = string(el)
		}
		return result_array, nil

	} else {
		return nil, errors.New("Internal reddan error, unexpected response type")
	}

}

// API Implementation starts here
// Example
// client.Get("keyname") -> "keyvalue", nil

// Key Commands
func (client *Client) Get(key string) (string, error) {
	return client.executeKeyCommand(buildCommand("GET", key))
}

func (client *Client) Set(key string, value string) (string, error) {
	return client.executeKeyCommand(buildCommand("SET", key, value))
}

func (client *Client) Ping() (string, error) {
	return client.executeKeyCommand(buildCommand("PING"))
}

func (client *Client) Del(keys ...string) (string, error) {
	return client.executeKeyCommand(buildCommand("DEL", keys...))
}

func (client *Client) Exists(key string) (bool, error) {
	return client.executeBoolCommand(buildCommand("EXISTS", key))
}

func (client *Client) Expire(key string, seconds int) (string, error) {
	return client.executeKeyCommand(buildCommand("EXPIRE", key, strconv.Itoa(seconds)))
}

func (client *Client) ExpireAt(key string, unixTime int) (string, error) {
	return client.executeKeyCommand(buildCommand("EXPIREAT", key, strconv.Itoa(unixTime)))
}

func (client *Client) Ttl(key string) (int, error) {
	return client.executeIntCommand(buildCommand("TTL", key))
}

func (client *Client) Keys(pattern string) ([]string, error) {
	return client.executeStringArrayCommand(buildCommand("KEYS", pattern))
}

func (client *Client) Move(key string, db int) (bool, error) {
	return client.executeBoolCommand(buildCommand("MOVE", key, strconv.Itoa(db)))
}

func (client *Client) Persist(key string) (bool ,error) {
	return client.executeBoolCommand(buildCommand("PERSIST", key))
}

func (client *Client) Pexpire(key string, miliseconds int) (bool, error){
	return client.executeBoolCommand(buildCommand("PEXPIRE", key, strconv.Itoa(miliseconds)))
}

func (client *Client) PexpireAt(key string, milisecondsTimestamp int) (bool, error){
	return client.executeBoolCommand(buildCommand("PEXPIREAT", key, strconv.Itoa(milisecondsTimestamp)))
}

func (client *Client) Pttl(key string) (int, error) {
	return client.executeIntCommand(buildCommand("PTTL", key))
}

func (client *Client) RandomKey() (string, error) {
	return client.executeKeyCommand(buildCommand("RANDOMKEY"))
}

func (client *Client) Rename(key string, newKey string) (string, error) {
	return client.executeKeyCommand(buildCommand("RENAME", key, newKey))
}

func (client *Client) RenameNx(key string, newKey string) (string, error) {
	return client.executeKeyCommand(buildCommand("RENAMENX", key, newKey))
}

func (client *Client) Type(key string) (string, error) {
	return client.executeKeyCommand(buildCommand("TYPE", key))
}

func (client *Client) Append(key string, value string) (int, error) {
	return client.executeIntCommand(buildCommand("APPEND", key, value))
}

func (client *Client) Strlen(key string) (int, error) {
	return client.executeIntCommand(buildCommand("STRLEN", key))
}

func (client *Client) Incr(key string) (int, error) {
	return client.executeIntCommand(buildCommand("INCR", key))
}

func (client *Client) Decr(key string) (int, error) {
	return client.executeIntCommand(buildCommand("DECR", key))
}

// List Commands
func (client *Client) Lpush(key string, val string) (int, error) {
	return client.executeIntCommand(buildCommand("LPUSH", key, val))
}

func (client *Client) LpushX(key string, val string) (int, error) {
	return client.executeIntCommand(buildCommand("LPUSHX", key, val))
}

func (client *Client) Rpush(key string, val string) (int, error) {
	return client.executeIntCommand(buildCommand("RPUSH", key, val))
}

func (client *Client) RpushX(key string, val string) (int, error) {
	return client.executeIntCommand(buildCommand("RPUSHX", key, val))
}

func (client *Client) Lpop(key string) (string, error) {
	return client.executeKeyCommand(buildCommand("LPOP", key))
}

func (client *Client) Rpop(key string) (string, error) {
	return client.executeKeyCommand(buildCommand("RPOP", key))
}

func (client *Client) BlPop(key string, args ...string) ([]string, error) {
	return client.executeStringArrayCommand(buildCommand("BLPOP", args...))
}

func (client *Client) BrPop(key string, args ...string) ([]string, error) {
	return client.executeStringArrayCommand(buildCommand("BRPOP", args...))
}

func (client *Client) Lrange(key string, from int, to int) ([]string, error) {
	return client.executeStringArrayCommand(buildCommand("LRANGE", key, strconv.Itoa(from), strconv.Itoa(to)))
}

func (client *Client) Llen(key string) (int, error) {
	return client.executeIntCommand(buildCommand("LLEN", key))
}

func (client *Client) Lindex(key string, pos int) (string, error) {
	return client.executeKeyCommand(buildCommand("LINDEX", key, strconv.Itoa(pos)))
}

func (client *Client) Lrem(list string, num int, key string) (int, error) {
	return client.executeIntCommand(buildCommand("LREM", list, strconv.Itoa(num), key))
}

func (client *Client) Lset(list string, pos int, key string) (string, error) {
	return client.executeKeyCommand(buildCommand("LSET", list, strconv.Itoa(pos), key))
}

func (client *Client) Ltrim(list string, from int, to int) (string ,error) {
	return client.executeKeyCommand(buildCommand("LTRIM", list, strconv.Itoa(from), strconv.Itoa(to)))
}



