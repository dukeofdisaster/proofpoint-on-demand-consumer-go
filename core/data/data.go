package data
/*
model the proofpoint on demand event structure
*/

type PODEvent struct {
	Connection struct {
		Country string `json:"country"`
		Helo string `json:"helo"`
		Host string `json:"host"`
		Ip string `json:"ip"`
		Protocol string `json:"protocol"`
		ResolveStatus string `json:"resolveStatus"`
		Sid string `json:"sid"`
		Tls  struct {
			Inbound struct {
				Cipher string `json:"cipher"`
				CipherBytes string `json:"cipherBytes"`
				Version string `json:"version"`
			}
		}
	}
	Envelope struct {
		From string `json:"from"`
		Rcpts []string `json:"rcpts"`
	}
	Filter struct {
		Actions string `json:"actions"`
		Delivered struct {
		}
		Dispostion string `json:"disposition"`
		DurationSecs float64 `json:"durationSecs"`
		// not sure on casing, this isn't in spec
		IsMsgInDigest bool `json:"isMsgInDigest"`
		Modules struct {
			Dkimv []DKIMItem `json:"dkimv"`
			Dmarc struct {
				AuthResults []AuthResultsItem `json:"authResults"`
				FilteredResult string `json:"filteredResult"`
				SrvId string `json:"srvid"`
			}
			Spam struct {
				Langs []string `json:"langs"`
				Scores struct {
					Classifiers struct {
						Adult int `json:"adult"`
						Bulk int `json:"bulk"`
						Impostor int `json:"impostor"`
						LowPriority int `json:"lowpriority"`
						Malware int `json:"malware"`
						Mlx int `json:"mlx"`
						MlxLog int `json:"mlxlog"`
						Phish int `json:"phish"`
						Spam int `json:"spam"`
						Suspect int `json:"suspect"`
					}
					Engine int `json:"engine"`
					Overall int `json:"overall"`
				}
				TriggeredClassifier string `json:"triggeredClassifier"`
				Version struct {
					Definitions string `json:"definitions"`
					Engine string `json:"engine"`
				}
			}
			Spf struct {
				Domain string `json:"domain"`
				Result string `json:"result"`
			}
			UrlDefense struct {
				Counts struct {
					NoRewriteIsEmail int `json:"noRewriteIsEmail"`
					Rewritten int `json:"rewritten"`
					Total int `json:"total"`
					Unique int `json:"unique"`
				}
				Version struct {
					Engine string `json:"engine"`
				}
			}
			ZeroHour struct {
				Score string `json:"score"`
			}
		}
		MsgSizeBytes int `json:"msgSizeBytes"`
		Qid string `json:"qid"`
		RouteDirection string `json:"routeDirection"`
		Suborgs struct {
			Rcpts []string `json:"rcpts"`
			Sender string `json:"sender"`
		}
		Verified struct {
			Rcpts []string `json:"rcpts"`
		}
	}
	Guid string `json:"guid"`
	Metadata struct {
		Origin struct {
			Data struct {
				Agent string `json:"agent"`
				Cid string `json:"cid"`
				Version string `json:"version"`
			}
		}
	}
	Msg struct {
		Header struct {
			From []string `json:"from"`
			MessageId []string `json:"message-id"`
			Subject []string `json:"subject"`
			To []string `json:"to"`
		}
		Lang string `json:"lang"`
		NormalizedHeader struct {
			From []string `json:"from"`
			MessageId []string `json:"message-id"`
			ReplyTo []string `json:"reply-to"`
			Subject []string `json:"subject"`
			To []string `json:"to"`
		}
		ParsedAddresses struct {
			From []string `json:"from"`
			FromDisplayNames []string `json:"fromDisplayNames"`
			To []string `json:"to"`
		}
		SizeBytes int `json:"sizeBytes"`
	}
	MsgParts []MessagePart `json:"msgParts"`
	Timestamp string `json:"ts"`
}
type AuthResultsItem struct {
	EmailIdentities struct {
		SmtpMailfrom string `json:"smtp.mailfrom"`	
	}
	Method string `json:"method"`
	Result string `json:"result"`
	PropSpec struct {
		HeaderD string `json:"header.d"`
		HeaderS string `json:"header.s"`
	}
}
type DKIMItem struct {
	Domain string `json:"domain"`
	Result string `json:"result"`
	selector string `json:"string"`
}

type MessagePart struct {
	DataBase64 string `json:"dataBase64"`
	DetectedCharset string `json:"detectedCharset"`
	DetectedExt string `json:"detectedExt"`
	DetectedMime string `json:"detectedMime"`
	DetectedName string `json:"detectedName"`
	// don't recall if these are actually int when they come across
	DetectedSizeBytes int `json:"detectedSizeBytes"`
	Disposition string `json:"disposition"`
	IsArchive bool `json:"isArchive"`
	IsCorrupted bool `json:"isCorrupted"`
	IsDeleted bool `json:"isDeleted"`
	IsProtected bool `json:"isProtected"`
	IsTimedOut bool `json:"isTimedOut"`
	IsVirtual bool `json:"isVirtual"`
	LabeledCharset string `json:"labeledCharset"`
	LabeledExt string `json:"labeledExt"`
	LabeledMime string `json:"labeledMime"`
	LabeledName string `json:"labeledName"`
	MD5 string `json:"md5"`
	SHA256 string `json:"sha256"`
	// don't recall if these are actually int when they come across
	SizeDecodedBytes int `json:"sizeDecodedBytes"`
	StructureId string `json:"structureId"`
	TextExtracted string `json:"textExtracted"`
	Urls []string `json:"urls"`
}

type Config struct {
	Endpoint string `yaml:"endpoint"`
	ApiKey string `yaml:"apikey"`
	Log struct {
		Path string `yaml:"path"`
	}
}
