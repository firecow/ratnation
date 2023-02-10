import wait from "wait-promise";
import {StateHandler} from "../state-handler.mjs";
import {KingContext} from "./king.mjs";
import {KingSyncer} from "./king-syncer.mjs";
import {KingRatholeManager} from "./king-rathole-manager.mjs";

interface KingShutdownHandlersOpts {
    context: KingContext;
    stateHandler: StateHandler;
    syncer: KingSyncer;
    ratholeManager: KingRatholeManager;
}

export function initKingShutdownHandlers ({context, stateHandler, syncer, ratholeManager}: KingShutdownHandlersOpts) {
    const logger = context.logger;
    const listener = async (signal: NodeJS.Signals) => {
        logger.info("Shutdown sequence initiated", {"service.type": "ratking"});
        context.shuttingDown = true;
        stateHandler.stop();
        syncer.stop();
        await syncer.tick();
        await wait.sleep(1000);
        ratholeManager.killProcesses(signal);
    };
    process.on("SIGINT", listener);
    process.on("SIGTERM", listener);
}
