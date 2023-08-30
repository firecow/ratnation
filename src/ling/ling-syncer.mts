import got from "got";
import {Ticker} from "../ticker.mjs";
import {LingContext} from "./ling.mjs";
import {to} from "../utils.mjs";

export class LingSyncer extends Ticker {

    private readonly context;

    constructor (context: LingContext) {
        super({interval: 1000, tick: async () => this.sync()});
        this.context = context;
    }

    private async sync () {
        const logger = this.context.logger;
        const [err, response] = await to(got(`${this.context.councilHost}/ling`, {
            method: "PUT",
            json: {
                ling_id: this.context.lingId,
                shutting_down: this.context.shuttingDown,
                ratholes: Array.from(this.context.config.ratholeMap.values()),
                ready_service_ids: this.context.readyServiceIds,
                preferred_location: "mylocation",
            },
        }));
        if (err || response.statusCode !== 200) {
            logger.error("Failed to sync with council", {
                "error.message": err?.message,
                "error.stack_trace": err?.stack,
                "service.type": "ratling",
            });
        }
    }
}
