import assert from "assert";

export class LingConfig {

    ratholeMap = new Map();
    socatMap = new Map();

    constructor(argv) {
        for (const socatArg of argv["socat"] ?? []) {
            const socatCnf = {};
            const pairs = socatArg.split(" ");
            for (const pair of pairs) {
                const key = pair.split("=")[0];
                socatCnf[key] = pair.split("=")[1];
            }
            socatCnf["bind_port"] = Number(socatCnf["bind_port"]);
            assert(socatCnf["bind_port"] != null, "--socat must have 'bind_port' field");
            assert(socatCnf["name"] != null, "--socat must have 'name' field");
            this.socatMap.set(socatCnf["name"], socatCnf);
        }

        for (const ratholeArg of argv["rathole"] ?? []) {
            const ratholeCnf = {};
            const pairs = ratholeArg.split(" ");
            for (const pair of pairs) {
                const key = pair.split("=")[0];
                ratholeCnf[key] = pair.split("=")[1];
            }
            assert(ratholeCnf["name"] != null, "--rathole must have 'name' field");
            assert(ratholeCnf["local_addr"] != null, "--rathole must have 'local_addr' field");
            this.ratholeMap.set(ratholeCnf["name"], ratholeCnf);
        }
    }
}
