import assert from "assert";
import crypto from "crypto";
import {LingArguments} from "../cmds/ling-cmd.js";

export interface LingRatholeConfig {
    name: string;
    local_addr: string;
}

export interface LingProxyConfig {
    bind_port: number;
    name: string;
}

export class LingConfig {

    readonly ratholeMap = new Map<string, LingRatholeConfig>();
    readonly proxyMap = new Map<string, LingProxyConfig>();
    readonly councilHost: string;
    readonly lingId: string;

    constructor (args: LingArguments) {
        this.councilHost = args.councilHost;
        this.lingId = args.lingId ?? crypto.randomUUID();

        for (const proxyArg of args.proxy ?? []) {
            const pairs: {[key: string]: string} = {};
            for (const pair of proxyArg.split(" ")) {
                const key = pair.split("=")[0];
                pairs[key] = pair.split("=")[1];
            }
            assert(pairs["bind_port"] != null, "--proxy must have 'bind_port' field");
            assert(pairs["name"] != null, "--proxy must have 'name' field");
            this.proxyMap.set(pairs["name"], {
                name: pairs["name"],
                bind_port: Number(pairs["bind_port"]),
            });
        }

        for (const ratholeArg of args.rathole ?? []) {
            const pairs: {[key: string]: string} = {};
            for (const pair of ratholeArg.split(" ")) {
                const key = pair.split("=")[0];
                pairs[key] = pair.split("=")[1];
            }
            assert(pairs["name"] != null, "--rathole must have 'name' field");
            assert(pairs["local_addr"] != null, "--rathole must have 'local_addr' field");
            this.ratholeMap.set(pairs["name"], {
                name: pairs["name"],
                local_addr: pairs["local_addr"],
            });
        }
    }
}
