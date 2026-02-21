package dispatcher

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type RedisDispatcher struct {
	rdb        *redis.Client
	ListPrefix string
	SetPrefix  string
}

func NewRedisDispatcher(rdb *redis.Client, listPrefix, setPrefix string) *RedisDispatcher {
	return &RedisDispatcher{
		rdb:        rdb,
		ListPrefix: listPrefix,
		SetPrefix:  setPrefix,
	}
}

func (d *RedisDispatcher) GetListKey(taskType string) string {
	return fmt.Sprintf("%s:%s", d.ListPrefix, taskType)
}

func (d *RedisDispatcher) GetSetKey(taskType string) string {
	return fmt.Sprintf("%s:%s", d.SetPrefix, taskType)
}

func (d *RedisDispatcher) Dispatch(ctx context.Context, taskType, taskId string) error {
	luaScript := `
		-- Redis Lua脚本：向list和set同时插入值（带条件判断）
		-- 参数说明：
		-- KEYS[1]: list的键名
		-- KEYS[2]: set的键名
		-- ARGV[1]: 要插入的值

		-- 向list中插入值（使用RPUSH，从列表尾部插入）
		local listResult = redis.call('RPUSH', KEYS[1], ARGV[1])

		-- 初始化set操作结果
		local setResult = 2

		-- 检查list插入是否成功（listResult应为数字，表示操作后列表长度）
		if type(listResult) == 'number' and listResult > 0 then
			-- 向set中插入值
			setResult = redis.call('SADD', KEYS[2], ARGV[1])
		end

		-- 返回结果：{list操作结果, set操作结果}
		return setResult`
	script := redis.NewScript(luaScript)
	errRet := script.Run(ctx, d.rdb, []string{d.GetListKey(taskType), d.GetSetKey(taskType)}, taskId)
	if errRet != nil {
		retCode, err := errRet.Int()
		if err == nil {
			if retCode == 2 {
				return fmt.Errorf("redis ret_code: %d", retCode)
			}
			return nil
		} else {
			return err
		}
	}
	return fmt.Errorf("unknown redis error")
}

func (d *RedisDispatcher) DispatchId(ctx context.Context, taskType string, taskId int64) error {
	taskIdStr := strconv.FormatInt(taskId, 10)
	return d.Dispatch(ctx, taskType, taskIdStr)
}

func (d *RedisDispatcher) GetTask(ctx context.Context, taskType string) (string, error) {
	luaScript := `
		-- Redis Lua脚本：从list中LPOP元素并同时从set中删除
		-- 参数说明：
		-- KEYS[1]: list的键名
		-- KEYS[2]: set的键名

		-- 从list中移除并获取元素（使用LPOP，从列表头部移除）
		local element = redis.call('LPOP', KEYS[1])

		-- 初始化set操作结果
		local setResult = 0

		-- 检查是否成功获取到元素
		if element then
			-- 从set中删除该元素
			setResult = redis.call('SREM', KEYS[2], element)
		end

		-- 返回结果：
		return {element, setResult}`
	script := redis.NewScript(luaScript)
	result, err := script.Run(ctx, d.rdb, []string{d.GetListKey(taskType), d.GetSetKey(taskType)}).Result()
	if err != nil {
		return "", err
	}
	resultArray, ok := result.([]interface{})
	if !ok {
		return "", fmt.Errorf("返回值格式错误: 期望[]interface{}，实际为%T", result)
	}

	// 检查返回值长度
	if len(resultArray) != 2 {
		return "", fmt.Errorf("返回值长度错误: 期望2个元素，实际为%d", len(resultArray))
	}
	var element string
	if resultArray[0] != nil {
		element, ok = resultArray[0].(string)
		if !ok {
			return "", fmt.Errorf("元素类型错误: 期望string，实际为%T", resultArray[0])
		}
	} else {
		fmt.Println("element is nil")
	}

	// 解析第二个返回值：set操作结果（SREM的返回值是整数）
	setResult, ok := resultArray[1].(int64)
	if !ok {
		// 注意：某些Redis客户端可能会返回float64类型，需要额外处理
		if floatVal, ok := resultArray[1].(float64); ok {
			setResult = int64(floatVal)
		} else {
			return "", fmt.Errorf("set操作结果类型错误: 期望int64，实际为%T", resultArray[1])
		}
	}
	fmt.Printf("element: %s, setResult: %d\n", element, setResult)
	return element, nil
}

func (d *RedisDispatcher) GetTaskId(ctx context.Context, taskType string) (int64, error) {
	idStr, err := d.GetTask(ctx, taskType)
	if err != nil {
		return 0, err
	}
	if len(idStr) == 0 {
		return 0, nil
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse redis string2int64 failed, err: %s", err.Error())
	}
	return id, nil
}

func (d *RedisDispatcher) ExistTask(ctx context.Context, taskType, taskId string) (bool, error) {
	exist, err := d.rdb.SIsMember(ctx, d.GetSetKey(taskType), taskId).Result()
	if err != nil {
		return false, err
	}
	return exist, nil
}

func (d *RedisDispatcher) ExistTaskId(ctx context.Context, taskType string, taskId int64) (bool, error) {
	taskIdStr := strconv.FormatInt(taskId, 10)
	exist, err := d.ExistTask(ctx, taskType, taskIdStr)
	if err != nil {
		return false, err
	}
	return exist, nil
}

// 返回：是否需要重新添加
func (d *RedisDispatcher) doReDispatch(ctx context.Context, taskType string, taskId string) (bool, error) {
	// 检测是否存在，如果不存在，就重新添加到redis
	exist, err := d.ExistTask(ctx, taskType, taskId)
	if err != nil {
		return false, err
	}
	if exist {
		// 存在，不需要重新添加
		return false, nil
	}
	/// 尽管获取任务是用lua脚本执行的，但极端情况下，还是存在这种情况：任务ID在set中，但list中已经取出的情况
	err = d.Dispatch(ctx, taskType, taskId)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *RedisDispatcher) ReDispatch(ctx context.Context, taskType string, taskIds []string) ([]string, error) {
	ids := make([]string, 0, len(taskIds))
	for _, id := range taskIds {
		isSucc, err := d.doReDispatch(ctx, taskType, id)
		if err != nil {
			return nil, err
		}
		if isSucc {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func (d *RedisDispatcher) ReDispatchIds(ctx context.Context, taskType string, taskIds []int64) ([]int64, error) {
	taskIdStrs := make([]string, 0, len(taskIds))
	for _, id := range taskIds {
		taskIdStrs = append(taskIdStrs, strconv.FormatInt(id, 10))
	}
	ids, err := d.ReDispatch(ctx, taskType, taskIdStrs)
	if err != nil {
		return nil, err
	}
	idsInt64 := make([]int64, 0, len(ids))
	for _, id := range ids {
		idInt64, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse redis string2int64 failed, err: %s", err.Error())
		}
		idsInt64 = append(idsInt64, idInt64)
	}
	return idsInt64, nil
}
