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
    const listener = async (signal: NodeJS.Signals) => {
        console.log("msg=\"king shutdown sequence initiated\" service.type=ratking");
        context.shuttingDown = true;
        void stateHandler.stop();
        await syncer.stop();
        await wait.sleep(1000);
        ratholeManager.killProcesses(signal);
    };
    process.on("SIGINT", listener);
    process.on("SIGTERM", listener);
}
