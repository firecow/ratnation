import got, {RequestError} from "got";
import {Ticker} from "../ticker.js";
import {KingContext} from "../contexts/king-context.js";
import {to} from "../utils.js";

export class KingSyncer extends Ticker {

    private readonly context;

    constructor (context: KingContext) {
        super({interval: 1000, tick: async () => this.sync()});

        this.context = context;
    }

    private async sync () {
        const config = this.context.config;
        const logger = this.context.logger;
        const [err, response] = await to(got(`${config.councilHost}/king`, {
            method: "PUT",
            json: {
                host: config.host,
                shutting_down: this.context.shuttingDown,
                ratholes: this.context.config.ratholes,
                ready_service_ids: this.context.readyServiceIds,
                location: this.context.config.location,
            },
        }));
        if (err || response.statusCode !== 200) {
            const body = err instanceof RequestError ? String(err.response?.body ?? "") : undefined;
            const message = body ? body.slice(4096) : err?.message;
            logger.error("Failed to sync with council", {
                "error.message": message,
                "error.stack_trace": err?.stack,
                "service.type": "ratking",
            });
        }
    }

}
