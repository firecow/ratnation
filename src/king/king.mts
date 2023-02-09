import {KingConfig} from "./king-config.mjs";
import {State, StateHandler} from "../state-handler.mjs";
import {KingRatholeManager} from "./king-rathole-manager.mjs";
import {KingSyncer} from "./king-syncer.mjs";
import {initKingShutdownHandlers} from "./king-shutdown.mjs";
import wait from "wait-promise";
import {ArgumentsCamelCase, Argv} from "yargs";

export interface KingArguments {
    "council-host": string;
    "rathole": string[];
    "host": string;
}

export class KingContext {
    config: KingConfig;
    state: State;
    readyServiceIds: string[];
    shuttingDown: boolean;
    councilHost: string;
    host: string;
    location: string;

    constructor (config: KingConfig, args: KingArguments) {
        this.config = config;
        this.state = {services: [], kings: [], lings: [], revision: 0};
        this.readyServiceIds = [];
        this.shuttingDown = false;
        this.host = args["host"];
        this.councilHost = args["council-host"];
        this.location = "mylocation"; // TODO: location from cli options
    }
}

export const command = "king";
export const description = "Start ratking";

export async function handler (args: ArgumentsCamelCase) {
    const config = new KingConfig(args as ArgumentsCamelCase<KingArguments>);
    const context = new KingContext(config, args as ArgumentsCamelCase<KingArguments>);
    const ratholeManager = new KingRatholeManager(context);
    const syncer = new KingSyncer(context);
    const stateHandler = new StateHandler({
        ...context,
        updatedFunc: async (state) => {
            context.state = state;

            await Promise.all([
                ratholeManager.stateChanged(),
            ]);
        },
    });
    initKingShutdownHandlers({context, stateHandler, syncer, ratholeManager});

    stateHandler.start();
    await wait.until(() => stateHandler.hasState());
    syncer.start();
    console.log("msg=\"king ready\" service.type=ratking");
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
    return yargs;
}
