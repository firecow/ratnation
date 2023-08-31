import {Logger} from "../logger.js";
import {LingConfig} from "../configs/ling-config.js";
import {State} from "../state-handler.js";

export class LingContext {
    logger: Logger;
    config: LingConfig;
    state: State;
    readyServiceIds: string[];
    shuttingDown: boolean;

    constructor (logger: Logger, config: LingConfig) {
        this.logger = logger;
        this.config = config;
        this.state = {services: [], kings: [], lings: [], revision: 0};
        this.readyServiceIds = [];
        this.shuttingDown = false;
    }
}