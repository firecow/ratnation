import assert from "assert";

export class LingConfig {

    ratholeMap = new Map();
    proxyMap = new Map();

    constructor(argv) {
        for (const proxyArg of argv["proxy"] ?? []) {
            const proxyCnf = {};
            const pairs = proxyArg.split(" ");
            for (const pair of pairs) {
                const key = pair.split("=")[0];
                proxyCnf[key] = pair.split("=")[1];
            }
            proxyCnf["bind_port"] = Number(proxyCnf["bind_port"]);
            assert(proxyCnf["bind_port"] != null, "--proxy must have 'bind_port' field");
            assert(proxyCnf["name"] != null, "--proxy must have 'name' field");
            this.proxyMap.set(proxyCnf["name"], proxyCnf);
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
