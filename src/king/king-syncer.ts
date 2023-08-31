import got from "got";
import {Ticker} from "../ticker.js";
import {KingContext} from "./king-cmd.js";
import {to} from "../utils.js";

export class KingSyncer extends Ticker {

    private readonly context;

    constructor (context: KingContext) {
        super({interval: 1000, tick: async () => this.sync()});
        this.context = context;
    }

    private async sync () {
        const logger = this.context.logger;
        const [err, response] = await to(got(`${this.context.councilHost}/king`, {
            method: "PUT",
            json: {
                host: this.context.host,
                shutting_down: this.context.shuttingDown,
                ratholes: this.context.config.ratholes,
                ready_service_ids: this.context.readyServiceIds,
                location: this.context.config.location,
            },
        }));
        if (err || response.statusCode !== 200) {
            logger.error("Failed to sync with council", {
                "error.message": err?.message,
                "error.stack_trace": err?.stack,
                "http.response.body": err?.response?.body?.slice(),
                "service.type": "ratking",
            });
        }
    }

}
