import waitFor from "p-wait-for";
import {ArgumentsCamelCase, Argv} from "yargs";
import {Logger} from "../logger.mjs";
import {State, StateHandler} from "../state-handler.mjs";
import {portsReachable} from "../utils.mjs";
import {KingConfig} from "./king-config.mjs";
import {KingRatholeManager} from "./king-rathole-manager.mjs";
import {initKingShutdownHandlers} from "./king-shutdown.mjs";
import {KingSyncer} from "./king-syncer.mjs";

export interface KingArguments {
    "council-host": string;
    "rathole": string[];
    "host": string;
    "location": string;
}

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

export const command = "king";
export const description = "Start ratking";

export async function handler (args: ArgumentsCamelCase) {
    const logger = new Logger();
    const config = new KingConfig(args as ArgumentsCamelCase<KingArguments>);
    await portsReachable(config.ratholes);
    const context = new KingContext(logger, config, args as ArgumentsCamelCase<KingArguments>);
    const ratholeManager = new KingRatholeManager(context);
    const syncer = new KingSyncer(context);
    const stateHandler = new StateHandler({
        ...context,
        stateChanged: async (state) => {
            context.state = state;
            await ratholeManager.stateChanged();
        },
    });
    initKingShutdownHandlers({context, stateHandler, syncer, ratholeManager});

    stateHandler.start();
    await waitFor(() => stateHandler.hasState());
    syncer.start();
    logger.info("Ready", {"service.type": "ratking"});
}

export function builder (yargs: Argv) {
    yargs.options("council-host", {
        type: "string",
        description: "Council host to syncronize from",
        default: "http://localhost:8080",
    });
    yargs.options("host", {
        type: "string",
        description: "Host (domain or ip)",
    });
    yargs.options("rathole", {
        type: "array",
        description: "Rathole servers to open",
    });
    yargs.options("location", {
        type: "string",
        description: "Location identifier",
        demandOption: true,
    });
    return yargs;
}
