import delay from "delay";
import {StateHandler} from "../state-handler.mjs";
import {LingRatholeManager} from "./ling-rathole-manager.mjs";
import {LingSyncer} from "./ling-syncer.mjs";
import {LingTraefikManager} from "./ling-traefik-manager.mjs";
import {LingContext} from "./ling.mjs";

interface LingShutdownHandlersOpts {
    context: LingContext;
    stateHandler: StateHandler;
    syncer: LingSyncer;
    traefikManager: LingTraefikManager;
    ratholeManager: LingRatholeManager;
}

export function initLingShutdownHandlers ({context, stateHandler, syncer, traefikManager, ratholeManager}: LingShutdownHandlersOpts) {
    const logger = context.logger;
    const listener = async (signal: NodeJS.Signals) => {
        if (context.shuttingDown) return;
        logger.info("Shutdown sequence initiated", {"service.type": "ratling"});
        context.shuttingDown = true;
        stateHandler.stop();
        syncer.stop();
        await syncer.tick().catch(() => logger.error("shutdown sync failed", {"service.type": "ratling"}));
        // Wait for kings to have noticed the ling shutdown state change.
        // TODO: We can do better that arbitrary sleep's
        await delay(750);
        await Promise.allSettled([
            ratholeManager.killProcesses(signal),
            traefikManager.killProcesses(signal),
        ]);
    };
    process.on("SIGINT", listener);
    process.on("SIGTERM", listener);
}
