import crypto from "crypto";
import wait from "wait-promise";
import {LingConfig} from "./ling-config.mjs";
import {LingSyncer} from "./ling-syncer.mjs";
import {State, StateHandler} from "../state-handler.mjs";
import {LingRatholeManager} from "./ling-rathole-manager.mjs";
import {initLingShutdownHandlers} from "./ling-shutdown.mjs";
import {LingTraefikManager} from "./ling-traefik-manager.mjs";
import {ArgumentsCamelCase, Argv} from "yargs";

export interface LingArguments {
    "council-host": string;
    "ling-id": string;
    "proxy": string[];
    "rathole": string[];
}

export class LingContext {
    config: LingConfig;
    state: State;
    readyServiceIds: string[];
    shuttingDown: boolean;
    councilHost: string;
    lingId: string;

    constructor (config: LingConfig, args: LingArguments) {
        this.config = config;
        this.state = {services: [], kings: [], lings: [], revision: 0};
        this.readyServiceIds = [];
        this.shuttingDown = false;
        this.councilHost = args["council-host"];
        this.lingId = args["ling-id"] ?? crypto.randomUUID();
    }
}

export const command = "ling";
export const description = "Start ratling";

export async function handler (args: ArgumentsCamelCase) {
    const config = new LingConfig(args as ArgumentsCamelCase<LingArguments>);
    const context = new LingContext(config, args as ArgumentsCamelCase<LingArguments>);
    const syncer = new LingSyncer(context);
    const traefikManager = new LingTraefikManager(context);
    const ratholeManager = new LingRatholeManager(context);
    const stateHandler = new StateHandler({
        ...context,
        updatedFunc: async (state) => {
            context.state = state;

            await Promise.all([
                traefikManager.stateChanged(),
                ratholeManager.stateChanged(),
            ]);
        },
    });
    initLingShutdownHandlers({context, stateHandler, syncer, traefikManager, ratholeManager});

    stateHandler.start();
    await wait.until(() => stateHandler.hasState());
    syncer.start();
    console.log("msg=\"ling ready\" service.type=ratling");
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
