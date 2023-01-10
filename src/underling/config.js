export class Config {

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
            socatCnf["port"] = Number(socatCnf["port"]);
            this.socatMap.set(socatCnf["name"], socatCnf);
        }

        for (const ratholeArg of argv["rathole"] ?? []) {
            const ratholeCnf = {};
            const pairs = ratholeArg.split(" ");
            for (const pair of pairs) {
                const key = pair.split("=")[0];
                ratholeCnf[key] = pair.split("=")[1];
            }
            this.ratholeMap.set(ratholeCnf["name"], ratholeCnf);
        }
    }
}
