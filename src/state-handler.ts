import got from "got";
import {Logger} from "./logger.js";
import {Ticker} from "./ticker.js";
import {to} from "./utils.js";

export interface StateKing {
    bind_port: number;
    host: string;
    ports: string;
    shutting_down: boolean;
    beat: number;
    location: string;
}

export interface StateLing {
    ling_id: string;
    shutting_down: boolean;
    beat: number;
}

export interface StateService {
    name: string;
    token: string;
    service_id: string;
    ling_id: string;
    preferred_location: string;

    ling_ready: boolean;
    king_ready: boolean;

    host: string | null;
    bind_port: number | null;
    remote_port: number | null;
}

export interface State {
    services: StateService[];
    kings: StateKing[];
    lings: StateLing[];
    revision: number;
}

interface StateHandlerOpts {
    logger: Logger;
    councilHost: string;
    stateChanged: (state: State) => Promise<void> | void;
}

export class StateHandler extends Ticker {

    private readonly logger;
    private readonly stateChanged;
    private readonly councilHost;
    private state: State | null = null;

    constructor ({logger, councilHost, stateChanged}: StateHandlerOpts) {
        super({interval: 500, tick: async () => await this.fetchState()});
        this.stateChanged = stateChanged;
        this.councilHost = councilHost;
        this.logger = logger;
    }

    hasState () {
        return this.state !== null;
    }

    async fetchState () {
        const logger = this.logger;
        const [err, response] = await to(got.get(`${this.councilHost}/state`));
        if (err || response.statusCode !== 200) {
            return logger.error("Failed to fetch state from council", {
                "error.message": err?.message,
                "error.stack_trace": err?.stack,
                "http.response.status_code": response?.statusCode,
            });
        }

        const newState = JSON.parse(response.body) as State;
        if (this.state === null || this.state.revision !== newState.revision) {
            this.state = newState;
            await this.stateChanged(newState);
        }
    }
}
