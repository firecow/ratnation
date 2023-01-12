import crypto from "crypto";
import wait from "wait-promise";
import {LingConfig} from "./ling-config.js";
import {LingSyncer} from "./ling-syncer.js";
import {StateHandler} from "../state-handler.js";
import {LingSocatManager} from "./ling-socat-manager.js";
import {LingRatholeManager} from "./ling-rathole-manager.js";

export const command = "ling";
export const description = "Start ratling";

export async function handler(argv) {
    const councilHost = argv["council-host"];
    const config = new LingConfig(argv);
    const uuid = argv["uuid"] ?? crypto.randomUUID();
    const context = {config, state: null, readyServices: [], councilHost, uuid};
    const configSyncer = new LingSyncer(context);
    const socatManager = new LingSocatManager(context);
    const ratholeManager = new LingRatholeManager(context);
    const stateHandler = new StateHandler({
        ...context,
        updatedFunc: (state) => {
            context.state = state;
            socatManager.doit();
            ratholeManager.doit();
        },
    });
    stateHandler.start();
    await wait.until(() => stateHandler.hasState());
    configSyncer.start();
    console.log("msg=\"ling ready\" service_type=ratling");
}

export function builder(yargs) {
    yargs.options("council-host", {
        type: "string",
        description: "Council host to syncronize from",
        default: "http://localhost:8080",
    });
    yargs.options("rathole", {
        type: "array",
        description: "Rathole clients to open if council state matches",
    });
    yargs.options("socat", {
        type: "array",
        description: "Socats to open based on config if council state matches",
    });
}
