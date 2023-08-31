import {Logger} from "../logger.js";
import {KingConfig} from "../configs/king-config.js";
import {State} from "../state-handler.js";
import {KingArguments} from "../cmds/king-cmd.js";

export class KingContext {
    logger: Logger;
    config: KingConfig;
    state: State;
    readyServiceIds: string[];
    shuttingDown: boolean;
    councilHost: string;
    host: string;

    constructor (logger: Logger, config: KingConfig, args: KingArguments) {
        this.logger = logger;
        this.config = config;
        this.state = {services: [], kings: [], lings: [], revision: 0};
        this.readyServiceIds = [];
        this.shuttingDown = false;
        this.host = args["host"];
        this.councilHost = args["council-host"];
    }
}