## Groxy - A simple Gopher-powered reverse proxy ##

Groxy is a simple reverse proxy written in Go, adapted to work with Artifactory. It can be used to proxy the Artifactory UI, as well as Docker V2 & V1 repositories.

## Installation ##

### Docker ###
```bash 
docker pull jfrog-support-docker-registry.bintray.io/jfrog-support/groxy
```

```bash
docker run -e DOCKER_MODE=true --name=groxy -d \
  -p 9010:9010 -p 9011:9011 -p 9012:9012 \
  -v ~/.groxy/config.json:/config.json \
  jfrog-support-docker-registry.bintray.io/jfrog-support/groxy
```

Monitor incoming traffic and upstream response by running `docker logs -f groxy`

## Manual ##

*The binary is compiled only for Darwin (OS X) at this time, it's recommended that you use that docker image instead*

1.Download the groxy executable from [Bintray](https://bintray.com/uriahl/generic/Groxy/view)

2.Create the ~/.groxy/config.json file:


```
{
   "ArtifactoryHost": "http://localhost:8080",
   "DefaultUIPort": "9010",
   "DefaultV1Port": "9011",
   "DefaultV2Port": "9012",
   "V1RepoKey": "docker-local-v1",
   "V2RepoKey": "docker-local-v2"
}
```

3. Add groxy to your path export PATH=$PATH:/path/to/groxy/executable

4.run the groxy executable:

```bash
groxy --conf=/Users/{you}/.groxy/config.json
```

Groxy will print incoming request and upstream response information, enabling easy HTTP debugging.
