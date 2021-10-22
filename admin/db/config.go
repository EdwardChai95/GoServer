package db

func GetConfigByKey(key string) map[string]string {

	configs, err := db.QueryString("select * from `config` WHERE `key` = '" + key +
		"' LIMIT 1")

	if err != nil {
		logger.Error(err)
	}
	if len(configs) > 0 {
		return configs[0]
	}

	return nil
}

func UpdateConfigs(data map[string][]string) error {
	for key, config := range data {
		if checkConfig := GetConfigByKey(key); checkConfig != nil {
			_, err := db.Exec("update `config` set `value` = ? where `key` = ?",
				config[0], key)
			if err != nil {
				logger.Printf("config update err: %v", err)
				return err
			}
			// update success
			continue
		}

		// if c, err := affected1.RowsAffected(); c == 0 && err == nil {
		_, err := db.Exec("INSERT INTO `config`(`key`, `value`) VALUES (?, ?)",
			key, config[0])
		if err != nil {
			logger.Printf("insert update err: %v", err)
			return err
		}
		// }
	}
	return nil
}

func GetConfigs() map[string]string {
	configs, err := db.QueryString("select * from `config`")

	if err != nil {
		logger.Error(err)
	}
	if len(configs) > 0 {
		configMap := map[string]string{}
		for _, c := range configs {
			configMap[c["key"]] = c["value"]
		}
		return configMap
	}

	return nil
}
