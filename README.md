## Groxy - A simple Gofer-powered reverse proxy ##

Groxy is a simple reverse proxy written in Go, adapted to work with Artifactory and the Bintray Development environment. It can be used to proxy the Artifactory and Bintray UIs, as well as Docker V2 & V1 repositories.

### Installation ###

1.Download the groxy executable from [Bintray](https://bintray.com/uriahl/generic/Groxy/view)

2.Create the ~/.groxy/config.json file:


```
#!JSON

{
   "ArtifactoryHost": "http://localhost:8080",
   "DefaultUIPort": "9015",
   "DefaultV1Port": "9011",
   "DefaultV2Port": "9012",
   "V1RepoKey": "docker-local-v1",
   "V2RepoKey": "docker-local-v2"
}

```

3.run the groxy executable. Groxy will print incoming request and upstream response information, enabling easy HTTP debugging.