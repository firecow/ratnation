import crypto from "crypto";
import wait from "wait-promise";
import {LingConfig} from "./ling-config.js";
import {LingSyncer} from "./ling-syncer.js";
import {StateHandler} from "../state-handler.js";
import {LingRatholeManager} from "./ling-rathole-manager.js";
import {initLingShutdownHandlers} from "./ling-shutdown.js";
import {LingTraefikManager} from "./ling-traefik-manager.js";

export const command = "ling";
export const description = "Start ratling";

export async function handler(argv) {
    const councilHost = argv["council-host"];
    const config = new LingConfig(argv);
    const lingId = argv["ling_id"] ?? crypto.randomUUID();
    const context = {config, state: null, readyServiceIds: [], shuttingDown: false, councilHost, lingId};
    const syncer = new LingSyncer(context);
    const traefikManager = new LingTraefikManager(context);
    const ratholeManager = new LingRatholeManager(context);
    const stateHandler = new StateHandler({
        ...context,
        updatedFunc: (state) => {
            context.state = state;
            traefikManager.stateChanged();
            ratholeManager.stateChanged();
        },
    });
    initLingShutdownHandlers({context, stateHandler, syncer, traefikManager, ratholeManager});

    stateHandler.start();
    await wait.until(() => stateHandler.hasState());
    syncer.start();
    console.log("msg=\"ling ready\" service.type=ratling");
}

export function builder(yargs) {
    yargs.options("council-host", {
        type: "string",
        description: "Council host to syncronize from",
        default: "http://localhost:8080",
    });
    yargs.options("ling_id", {
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
}
