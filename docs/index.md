# プラグインAPI仕様書
<a name="top"></a>

## インデックス
- [API仕様](#API仕様)

  - [proto/session.proto](#proto/session.proto)
      - [DeleteRequest](#proto.DeleteRequest)
      - [Empty](#proto.Empty)
      - [GetRequest](#proto.GetRequest)
      - [GetResponse](#proto.GetResponse)
      - [PutRequest](#proto.PutRequest)
      - [SettingRequest](#proto.SettingRequest)
  
  
  
      - [Session](#proto.Session)
  

- [スカラー値型](#スカラー値型)

## API仕様


<a name="proto/session.proto"></a>
<p align="right"><a href="#top">Top</a></p>

### proto/session.proto



<a name="proto.DeleteRequest"></a>

#### DeleteRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |






<a name="proto.Empty"></a>

#### Empty







<a name="proto.GetRequest"></a>

#### GetRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |






<a name="proto.GetResponse"></a>

#### GetResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [string](#string) |  |  |






<a name="proto.PutRequest"></a>

#### PutRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="proto.SettingRequest"></a>

#### SettingRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| config | [bytes](#bytes) |  | repeated string endpoints = 1; int32 cacheTime = 2; string userName = 3; string password = 4; |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="proto.Session"></a>

#### Session


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Init | [SettingRequest](#proto.SettingRequest) | [Empty](#proto.Empty) |  |
| Get | [GetRequest](#proto.GetRequest) | [GetResponse](#proto.GetResponse) |  |
| Put | [PutRequest](#proto.PutRequest) | [Empty](#proto.Empty) |  |
| Delete | [DeleteRequest](#proto.DeleteRequest) | [Empty](#proto.Empty) |  |
| Close | [Empty](#proto.Empty) | [Empty](#proto.Empty) |  |

 <!-- end services -->



## スカラー値型

| .proto Type | Notes | Go Type | C++ Type | Java Type | Python Type |
| ----------- | ----- | -------- | -------- | --------- | ----------- |
| <a name="double" /> double |  | float64 | double | double | float |
| <a name="float" /> float |  | float32 | float | float | float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int32 | int | int |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | int64 | long | int/long |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | uint32 | int | int/long |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | uint64 | long | int/long |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int32 | int | int |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | int64 | long | int/long |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | uint32 | int | int |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | uint64 | long | int/long |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int32 | int | int |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | int64 | long | int/long |
| <a name="bool" /> bool |  | bool | bool | boolean | boolean |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | string | String | str/unicode |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | []byte | string | ByteString | str |
