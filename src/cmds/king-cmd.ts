import waitFor from "p-wait-for";
import {ArgumentsCamelCase, Argv} from "yargs";
import {Logger} from "../logger.js";
import {StateHandler} from "../state-handler.js";
import {portsReachable} from "../utils.js";
import {KingConfig} from "../configs/king-config.js";
import {KingRatholeManager} from "../managers/king-rathole-manager.js";
import {initKingShutdownHandlers} from "../shutdown/king-shutdown.js";
import {KingSyncer} from "../tickers/king-syncer.js";
import {KingContext} from "../contexts/king-context.js";

export interface KingArguments {
    councilHost: string;
    rathole: string[];
    host: string;
    location: string;
}

export const command = "king";
export const description = "Start ratking";

export async function handler (args: ArgumentsCamelCase) {
    const logger = new Logger();
    const config = new KingConfig(args as ArgumentsCamelCase<KingArguments>);
    await portsReachable(config.ratholes);
    const context = new KingContext(logger, config);
    const ratholeManager = new KingRatholeManager(context);
    const syncer = new KingSyncer(context);
    const stateHandler = new StateHandler({
        ...context,
        councilHost: config.councilHost,
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
