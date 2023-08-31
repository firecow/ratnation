import {Logger} from "../logger.js";
import {LingConfig} from "../configs/ling-config.js";
import {State} from "../state-handler.js";
import crypto from "crypto";
import {LingArguments} from "../cmds/ling-cmd.js";

export class LingContext {
    logger: Logger;
    config: LingConfig;
    state: State;
    readyServiceIds: string[];
    shuttingDown: boolean;
    councilHost: string;
    lingId: string;

    constructor (logger: Logger, config: LingConfig, args: LingArguments) {
        this.logger = logger;
        this.config = config;
        this.state = {services: [], kings: [], lings: [], revision: 0};
        this.readyServiceIds = [];
        this.shuttingDown = false;
        this.councilHost = args["council-host"];
        this.lingId = args["ling-id"] ?? crypto.randomUUID();
    }
}