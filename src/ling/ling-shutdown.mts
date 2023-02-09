import wait from "wait-promise";
import {LingTraefikManager} from "./ling-traefik-manager.mjs";
import {LingSyncer} from "./ling-syncer.mjs";
import {StateHandler} from "../state-handler.mjs";
import {LingContext} from "./ling.mjs";
import {LingRatholeManager} from "./ling-rathole-manager.mjs";

interface LingShutdownHandlersOpts {
    context: LingContext;
    stateHandler: StateHandler;
    syncer: LingSyncer;
    traefikManager: LingTraefikManager;
    ratholeManager: LingRatholeManager;
}

export function initLingShutdownHandlers ({context, stateHandler, syncer, traefikManager, ratholeManager}: LingShutdownHandlersOpts) {
    const listener = async (signal: NodeJS.Signals) => {
        console.log("message=\"ling shutdown sequence initiated\" service.type=ratling");
        context.shuttingDown = true;
        void stateHandler.stop();
        await syncer.stop();
        await wait.sleep(1000);
        ratholeManager.killProcesses(signal);
        traefikManager.killProcesses(signal);
    };
    process.on("SIGINT", listener);
    process.on("SIGTERM", listener);
}
