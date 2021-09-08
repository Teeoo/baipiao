local logger = require("logger")

function Process(data)
    logger.info("我是限时活动Process", {})
    return 1
end