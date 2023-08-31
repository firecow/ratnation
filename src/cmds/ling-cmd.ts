import waitFor from "p-wait-for";
import {ArgumentsCamelCase, Argv} from "yargs";
import {Logger} from "../logger.js";
import {StateHandler} from "../state-handler.js";
import {portsReachable} from "../utils.js";
import {LingConfig} from "../configs/ling-config.js";
import {LingRatholeManager} from "../managers/ling-rathole-manager.js";
import {initLingShutdownHandlers} from "../shutdown/ling-shutdown.js";
import {LingSyncer} from "../tickers/ling-syncer.js";
import {LingTraefikManager} from "../managers/ling-traefik-manager.js";
import {LingContext} from "../contexts/ling-context.js";

export interface LingArguments {
    "council-host": string;
    "ling-id": string;
    "proxy": string[];
    "rathole": string[];
}

export const command = "ling";
export const description = "Start ratling";

export async function handler (args: ArgumentsCamelCase) {
    const logger = new Logger();
    const config = new LingConfig(args as ArgumentsCamelCase<LingArguments>);
    await portsReachable(config.proxyMap.values());
    const context = new LingContext(logger, config, args as ArgumentsCamelCase<LingArguments>);
    const syncer = new LingSyncer(context);
    const traefikManager = new LingTraefikManager(context);
    const ratholeManager = new LingRatholeManager(context);
    const stateHandler = new StateHandler({
        ...context,
        stateChanged: async (state) => {
            context.state = state;

            await Promise.all([
                ratholeManager.stateChanged(),
                traefikManager.stateChanged(),
            ]);
        },
    });
    initLingShutdownHandlers({context, stateHandler, syncer, traefikManager, ratholeManager});

    stateHandler.start();
    await waitFor(() => stateHandler.hasState());
    syncer.start();
    logger.info("Ready", {"service.type": "ratling"});
}

export function builder (yargs: Argv) {
    yargs.options("council-host", {
        type: "string",
        description: "Council host to syncronize from",
        default: "http://localhost:8080",
    });
    yargs.options("ling-id", {
        type: "string",
        description: "Unique id of this specific ling instance",
        optional: true,
    });
    yargs.options("rathole", {
        type: "array",
        description: "Rathole clients to open",
    });
    yargs.options("proxy", {
        type: "array",
        description: "Traefik proxies to open",
    });
    return yargs;
}
