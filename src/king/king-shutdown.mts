import delay from "delay";
import {StateHandler} from "../state-handler.mjs";
import {KingRatholeManager} from "./king-rathole-manager.mjs";
import {KingSyncer} from "./king-syncer.mjs";
import {KingContext} from "./king.mjs";

interface KingShutdownHandlersOpts {
    context: KingContext;
    stateHandler: StateHandler;
    syncer: KingSyncer;
    ratholeManager: KingRatholeManager;
}

export function initKingShutdownHandlers ({context, stateHandler, syncer, ratholeManager}: KingShutdownHandlersOpts) {
    const logger = context.logger;
    const listener = async (signal: NodeJS.Signals) => {
        if (context.shuttingDown) return;
        logger.info("Shutdown sequence initiated", {"service.type": "ratking"});
        context.shuttingDown = true;
        stateHandler.stop();
        syncer.stop();
        await syncer.tick().catch(() => logger.error("shutdown sync failed", {"service.type": "ratking"}));
        // Wait for lings to have noticed the king shutdown state change.
        // TODO: We can do better that arbitrary sleep's
        await delay(1000);
        await ratholeManager.killProcesses(signal);
    };
    process.on("SIGINT", listener);
    process.on("SIGTERM", listener);
}
