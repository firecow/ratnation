import assert from "assert";
import {KingArguments} from "../cmds/king-cmd.js";

export interface KingRatholeConfig {
    ports: string;
    bind_port: number;
}

export class KingConfig {

    ratholes: KingRatholeConfig[] = [];
    location: string;
    host: string;
    councilHost: string;

    constructor (args: KingArguments) {
        this.location = args.location;
        this.host = args.host;
        this.councilHost = args.councilHost;

        for (const ratholeArg of args.rathole ?? []) {
            const pairs: {[key: string]: string} = {};
            for (const pair of ratholeArg.split(" ")) {
                const key = pair.split("=")[0];
                pairs[key] = pair.split("=")[1];
            }
            assert(pairs["bind_port"] != null, "--rathole must have 'bind_port' field");
            assert(pairs["ports"] != null, "--rathole must have 'name' field");
            this.ratholes.push({
                ports: pairs["ports"],
                bind_port: Number(pairs["bind_port"]),
            });
        }

        assert(this.ratholes.length > 0, "One --rathole must be specified");
    }
}
