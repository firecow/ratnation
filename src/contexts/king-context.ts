import {Logger} from "../logger.js";
import {KingConfig} from "../configs/king-config.js";
import {State} from "../state-handler.js";

export class KingContext {
    logger: Logger;
    config: KingConfig;
    state: State;
    readyServiceIds: string[];
    shuttingDown: boolean;

    constructor (logger: Logger, config: KingConfig) {
        this.logger = logger;
        this.config = config;
        this.state = {services: [], kings: [], lings: [], revision: 0};
        this.readyServiceIds = [];
        this.shuttingDown = false;
    }
}