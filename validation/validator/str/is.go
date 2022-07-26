// The MIT License (MIT)
//
// Copyright (c) 2014-2020 Alex Saskevich
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package str is ported from github.com/asaskevich/govalidator@v11.0.1
// because govalidator does not conform with Go Module v2+.
package str

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	maxURLRuneCount    = 2083
	minURLRuneCount    = 3
	rfc3339WithoutZone = "2006-01-02T15:04:05"
)

var whiteSpacesAndMinus = regexp.MustCompile(`[\s-]+`)

// IsEmail checks if the string is an email.
//
// TODO: uppercase letters are not supported
func IsEmail(str string) bool { return rxEmail.MatchString(str) }

// IsExistingEmail checks if the string is an email of existing domain
func IsExistingEmail(email string) bool {
	if len(email) < 6 || len(email) > 254 {
		return false
	}

	at := strings.LastIndex(email, "@")
	if at <= 0 || at > len(email)-3 {
		return false
	}

	user := email[:at]
	host := email[at+1:]
	if len(user) > 64 {
		return false
	}

	switch host {
	case "localhost", "example.com":
		return true
	}

	if userDotRegexp.MatchString(user) ||
		!userRegexp.MatchString(user) ||
		!hostRegexp.MatchString(host) {
		return false
	}

	if _, err := net.LookupMX(host); err != nil {
		if _, err := net.LookupIP(host); err != nil {
			return false
		}
	}

	return true
}

// IsURL checks if the string is an URL.
func IsURL(str string) bool {
	if str == "" || utf8.RuneCountInString(str) >= maxURLRuneCount ||
		len(str) <= minURLRuneCount || strings.HasPrefix(str, ".") {
		return false
	}

	strTemp := str
	if strings.Contains(str, ":") && !strings.Contains(str, "://") {
		// support no indicated urlscheme but with colon for port number
		// http:// is appended so url.Parse will succeed, strTemp used
		// so it does not impact rxURL.MatchString
		strTemp = "http://" + str
	}

	u, err := url.Parse(strTemp)
	if err != nil {
		return false
	}

	if strings.HasPrefix(u.Host, ".") {
		return false
	}
	if u.Host == "" && (u.Path != "" && !strings.Contains(u.Path, ".")) {
		return false
	}

	return rxURL.MatchString(str)
}

// IsRequestURL checks if the string rawurl, assuming it was received
// in an HTTP request, is a valid URL confirm to RFC 3986.
func IsRequestURL(rawurl string) bool {
	url, err := url.ParseRequestURI(rawurl)
	if err != nil {
		return false //Couldn't even parse the rawurl
	}

	if len(url.Scheme) == 0 {
		return false //No Scheme found
	}
	return true
}

// IsRequestURI checks if the string rawurl, assuming it was received
// in an HTTP request, is an absolute URI or an absolute path.
func IsRequestURI(rawurl string) bool {
	_, err := url.ParseRequestURI(rawurl)
	return err == nil
}

// IsAlpha checks if the string contains only letters (a-zA-Z).
//
// Empty string is valid.
func IsAlpha(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxAlpha.MatchString(str)
}

// IsUTFLetter checks if the string contains only unicode letter characters.
// Similar to IsAlpha but for all languages.
//
// Empty string is valid.
func IsUTFLetter(str string) bool {
	if IsNull(str) {
		return true
	}

	for _, c := range str {
		if !unicode.IsLetter(c) {
			return false
		}
	}

	return true
}

// IsAlphanumeric checks if the string contains only letters and numbers.
//
// Empty string is valid.
func IsAlphanumeric(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxAlphanumeric.MatchString(str)
}

// IsUTFLetterNumeric checks if the string contains only unicode letters and numbers.
//
// Empty string is valid.
func IsUTFLetterNumeric(str string) bool {
	if IsNull(str) {
		return true
	}

	for _, c := range str {
		if !unicode.IsLetter(c) && !unicode.IsNumber(c) { //letters && numbers are ok
			return false
		}
	}

	return true
}

// IsNumeric checks if the string contains only numbers.
//
// Empty string is valid.
func IsNumeric(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxNumeric.MatchString(str)
}

// IsUTFNumeric checks if the string contains only unicode numbers of any kind.
// Numbers can be 0-9 but also Fractions ¾,Roman Ⅸ and Hangzhou 〩.
//
// Empty string is valid.
func IsUTFNumeric(str string) bool {
	if IsNull(str) {
		return true
	}
	if strings.IndexAny(str, "+-") > 0 {
		return false
	}

	if len(str) > 1 {
		str = strings.TrimPrefix(str, "-")
		str = strings.TrimPrefix(str, "+")
	}
	for _, c := range str {
		if !unicode.IsNumber(c) { //numbers && minus sign are ok
			return false
		}
	}

	return true
}

// IsUTFDigit checks if the string contains only unicode radix-10 decimal digits.
//
// Empty string is valid.
func IsUTFDigit(str string) bool {
	if IsNull(str) {
		return true
	}
	if strings.IndexAny(str, "+-") > 0 {
		return false
	}

	if len(str) > 1 {
		str = strings.TrimPrefix(str, "-")
		str = strings.TrimPrefix(str, "+")
	}
	for _, c := range str {
		if !unicode.IsDigit(c) { //digits && minus sign are ok
			return false
		}
	}

	return true
}

// IsHexadecimal checks if the string is a hexadecimal number.
func IsHexadecimal(str string) bool { return rxHexadecimal.MatchString(str) }

// IsHexcolor checks if the string is a hexadecimal color.
func IsHexcolor(str string) bool { return rxHexcolor.MatchString(str) }

// IsRGBcolor checks if the string is a valid RGB color in form rgb(RRR, GGG, BBB).
func IsRGBcolor(str string) bool { return rxRGBcolor.MatchString(str) }

// IsLowerCase checks if the string is lowercase.
//
// Empty string is valid.
func IsLowerCase(str string) bool {
	if IsNull(str) {
		return true
	}
	return str == strings.ToLower(str)
}

// IsUpperCase checks if the string is uppercase.
//
// Empty string is valid.
func IsUpperCase(str string) bool {
	if IsNull(str) {
		return true
	}
	return str == strings.ToUpper(str)
}

// HasLowerCase checks if the string contains at least 1 lowercase.
//
// Empty string is valid.
func HasLowerCase(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxHasLowerCase.MatchString(str)
}

// HasUpperCase checks if the string contains as least 1 uppercase.
//
// Empty string is valid.
func HasUpperCase(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxHasUpperCase.MatchString(str)
}

// IsInt checks if the string is an integer.
//
// Empty string is valid.
func IsInt(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxInt.MatchString(str)
}

// IsFloat checks if the string is a float.
func IsFloat(str string) bool { return str != "" && rxFloat.MatchString(str) }

// IsNull checks if the string is null.
func IsNull(str string) bool { return len(str) == 0 }

// IsNotNull checks if the string is not null.
func IsNotNull(str string) bool { return !IsNull(str) }

// HasWhitespaceOnly checks the string only contains whitespace.
func HasWhitespaceOnly(str string) bool {
	return len(str) > 0 && rxHasWhitespaceOnly.MatchString(str)
}

// HasWhitespace checks if the string contains any whitespace.
func HasWhitespace(str string) bool {
	return len(str) > 0 && rxHasWhitespace.MatchString(str)
}

// IsByteLength checks if the string's length (in bytes) falls in a range.
func IsByteLength(str string, min, max int) bool {
	return len(str) >= min && len(str) <= max
}

// IsUUIDv3 checks if the string is a UUID version 3.
func IsUUIDv3(str string) bool { return rxUUID3.MatchString(str) }

// IsUUIDv4 checks if the string is a UUID version 4.
func IsUUIDv4(str string) bool { return rxUUID4.MatchString(str) }

// IsUUIDv5 checks if the string is a UUID version 5.
func IsUUIDv5(str string) bool { return rxUUID5.MatchString(str) }

// IsUUID checks if the string is a UUID (version 3, 4 or 5).
func IsUUID(str string) bool { return rxUUID.MatchString(str) }

// IsISBN10 checks if the string is an ISBN version 10.
func IsISBN10(str string) bool { return IsISBN(str, 10) }

// IsISBN13 checks if the string is an ISBN version 13.
func IsISBN13(str string) bool { return IsISBN(str, 13) }

// IsISBN checks if the string is an ISBN (version 10 or 13).
//
// If version value is not equal to 10 or 13, it will be checks both variants.
func IsISBN(str string, version int) bool {
	sanitized := whiteSpacesAndMinus.ReplaceAllString(str, "")
	var checksum int32
	var i int32
	if version == 10 {
		if !rxISBN10.MatchString(sanitized) {
			return false
		}

		for i = 0; i < 9; i++ {
			checksum += (i + 1) * int32(sanitized[i]-'0')
		}

		if sanitized[9] == 'X' {
			checksum += 10 * 10
		} else {
			checksum += 10 * int32(sanitized[9]-'0')
		}

		if checksum%11 == 0 {
			return true
		}
		return false
	} else if version == 13 {
		if !rxISBN13.MatchString(sanitized) {
			return false
		}

		factor := []int32{1, 3}
		for i = 0; i < 12; i++ {
			checksum += factor[i%2] * int32(sanitized[i]-'0')
		}

		return (int32(sanitized[12]-'0'))-((10-(checksum%10))%10) == 0
	}

	return IsISBN(str, 10) || IsISBN(str, 13)
}

// IsJSON checks if the string is valid JSON (note: uses json.Unmarshal).
func IsJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

// IsMultibyte checks if the string contains one or more multibyte chars.
//
// Empty string is valid.
func IsMultibyte(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxMultibyte.MatchString(str)
}

// IsASCII checks if the string contains ASCII chars only.
//
// Empty string is valid.
func IsASCII(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxASCII.MatchString(str)
}

// IsPrintableASCII checks if the string contains printable ASCII chars only.
//
// Empty string is valid.
func IsPrintableASCII(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxPrintableASCII.MatchString(str)
}

// IsFullWidth checks if the string contains any full-width chars.
//
// Empty string is valid.
func IsFullWidth(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxFullWidth.MatchString(str)
}

// IsHalfWidth checks if the string contains any half-width chars.
//
// Empty string is valid.
func IsHalfWidth(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxHalfWidth.MatchString(str)
}

// IsVariableWidth checks if the string contains a mixture of full and half-width chars.
//
// Empty string is valid.
func IsVariableWidth(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxHalfWidth.MatchString(str) && rxFullWidth.MatchString(str)
}

// IsBase64 checks if a string is base64 encoded.
func IsBase64(str string) bool { return rxBase64.MatchString(str) }

// IsDataURI checks if a string is base64 encoded data URI such as an image.
func IsDataURI(str string) bool {
	dataURI := strings.Split(str, ",")
	if !rxDataURI.MatchString(dataURI[0]) {
		return false
	}
	return IsBase64(dataURI[1])
}

// IsMagnetURI checks if a string is valid magnet URI.
func IsMagnetURI(str string) bool { return rxMagnetURI.MatchString(str) }

// IsDNSName will validate the given string as a DNS name.
func IsDNSName(str string) bool {
	if str == "" || len(strings.Replace(str, ".", "", -1)) > 255 {
		// constraints already violated
		return false
	}
	return !IsIP(str) && rxDNSName.MatchString(str)
}

// IsHash checks if a string is a hash of type algorithm.
//
// Algorithm is one of
//   - 'md4', 'md5'
//   - 'crc32', 'crc32b'
//   - 'sha1', 'sha256', 'sha384', 'sha512'
//   - 'tiger128', 'tiger160', 'tiger192'
//   - 'ripemd128', 'ripemd160'
func IsHash(str string, algorithm string) bool {
	var len string
	algo := strings.ToLower(algorithm)

	if algo == "crc32" || algo == "crc32b" {
		len = "8"
	} else if algo == "md5" || algo == "md4" || algo == "ripemd128" || algo == "tiger128" {
		len = "32"
	} else if algo == "sha1" || algo == "ripemd160" || algo == "tiger160" {
		len = "40"
	} else if algo == "tiger192" {
		len = "48"
	} else if algo == "sha3-224" {
		len = "56"
	} else if algo == "sha256" || algo == "sha3-256" {
		len = "64"
	} else if algo == "sha384" || algo == "sha3-384" {
		len = "96"
	} else if algo == "sha512" || algo == "sha3-512" {
		len = "128"
	} else {
		return false
	}

	match, _ := regexp.MatchString("^[a-f0-9]{"+len+"}$", str)
	return match
}

// IsSHA3224 checks is a string is a SHA3-224 hash.
//
// Alias for `IsHash(str, "sha3-224")`
func IsSHA3224(str string) bool { return IsHash(str, "sha3-224") }

// IsSHA3256 checks is a string is a SHA3-256 hash.
//
// Alias for `IsHash(str, "sha3-256")`
func IsSHA3256(str string) bool { return IsHash(str, "sha3-256") }

// IsSHA3384 checks is a string is a SHA3-384 hash.
//
// Alias for `IsHash(str, "sha3-384")`
func IsSHA3384(str string) bool { return IsHash(str, "sha3-384") }

// IsSHA3512 checks is a string is a SHA3-512 hash.
//
// Alias for `IsHash(str, "sha3-512")`
func IsSHA3512(str string) bool { return IsHash(str, "sha3-512") }

// IsSHA512 checks is a string is a SHA512 hash.
//
// Alias for `IsHash(str, "sha512")`
func IsSHA512(str string) bool { return IsHash(str, "sha512") }

// IsSHA384 checks is a string is a SHA384 hash.
//
// Alias for `IsHash(str, "sha384")`
func IsSHA384(str string) bool { return IsHash(str, "sha384") }

// IsSHA256 checks is a string is a SHA256 hash.
//
// Alias for `IsHash(str, "sha256")`
func IsSHA256(str string) bool { return IsHash(str, "sha256") }

// IsTiger192 checks is a string is a Tiger192 hash.
//
// Alias for `IsHash(str, "tiger192")`
func IsTiger192(str string) bool { return IsHash(str, "tiger192") }

// IsTiger160 checks is a string is a Tiger160 hash.
//
// Alias for `IsHash(str, "tiger160")`
func IsTiger160(str string) bool { return IsHash(str, "tiger160") }

// IsRipeMD160 checks is a string is a RipeMD160 hash.
//
// Alias for `IsHash(str, "ripemd160")`
func IsRipeMD160(str string) bool { return IsHash(str, "ripemd160") }

// IsSHA1 checks is a string is a SHA-1 hash.
//
// Alias for `IsHash(str, "sha1")`
func IsSHA1(str string) bool { return IsHash(str, "sha1") }

// IsTiger128 checks is a string is a Tiger128 hash.
//
// Alias for `IsHash(str, "tiger128")`
func IsTiger128(str string) bool { return IsHash(str, "tiger128") }

// IsRipeMD128 checks is a string is a RipeMD128 hash.
//
// Alias for `IsHash(str, "ripemd128")`
func IsRipeMD128(str string) bool { return IsHash(str, "ripemd128") }

// IsCRC32 checks is a string is a CRC32 hash.
//
// Alias for `IsHash(str, "crc32")`
func IsCRC32(str string) bool { return IsHash(str, "crc32") }

// IsCRC32b checks is a string is a CRC32b hash.
//
// Alias for `IsHash(str, "crc32b")`
func IsCRC32b(str string) bool { return IsHash(str, "crc32b") }

// IsMD5 checks is a string is a MD5 hash.
//
// Alias for `IsHash(str, "md5")`
func IsMD5(str string) bool { return IsHash(str, "md5") }

// IsMD4 checks is a string is a MD4 hash.
//
// Alias for `IsHash(str, "md4")`
func IsMD4(str string) bool { return IsHash(str, "md4") }

// IsDialString validates the given string for usage with the various Dial() functions.
func IsDialString(str string) bool {
	if h, p, err := net.SplitHostPort(str); err == nil && h != "" && p != "" &&
		(IsDNSName(h) || IsIP(h)) && IsPort(p) {
		return true
	}
	return false
}

// IsIP checks if a string is either IP version 4 or 6.
//
// Alias for `net.ParseIP`
func IsIP(str string) bool { return net.ParseIP(str) != nil }

// IsPort checks if a string represents a valid port.
func IsPort(str string) bool {
	if i, err := strconv.Atoi(str); err == nil && i > 0 && i < 65536 {
		return true
	}
	return false
}

// IsIPv4 checks if the string is an IP version 4.
func IsIPv4(str string) bool {
	ip := net.ParseIP(str)
	return ip != nil && strings.Contains(str, ".")
}

// IsIPv6 checks if the string is an IP version 6.
func IsIPv6(str string) bool {
	ip := net.ParseIP(str)
	return ip != nil && strings.Contains(str, ":")
}

// IsCIDR checks if the string is an valid CIDR notiation (IPV4 & IPV6).
func IsCIDR(str string) bool {
	_, _, err := net.ParseCIDR(str)
	return err == nil
}

// IsMAC checks if a string is valid MAC address.
//
// Possible MAC formats:
//   - 01:23:45:67:89:ab
//   - 01:23:45:67:89:ab:cd:ef
//   - 01-23-45-67-89-ab
//   - 01-23-45-67-89-ab-cd-ef
//   - 0123.4567.89ab
//   - 0123.4567.89ab.cdef
func IsMAC(str string) bool {
	_, err := net.ParseMAC(str)
	return err == nil
}

// IsHost checks if the string is a valid IP (both v4 and v6) or a valid DNS name.
func IsHost(str string) bool { return IsIP(str) || IsDNSName(str) }

// IsMongoID checks if the string is a valid hex-encoded representation
// of a MongoDB ObjectId.
func IsMongoID(str string) bool {
	return rxHexadecimal.MatchString(str) && (len(str) == 24)
}

// IsLatitude checks if a string is valid latitude.
func IsLatitude(str string) bool { return rxLatitude.MatchString(str) }

// IsLongitude checks if a string is valid longitude.
func IsLongitude(str string) bool { return rxLongitude.MatchString(str) }

// IsIMEI checks if a string is valid IMEI
func IsIMEI(str string) bool { return rxIMEI.MatchString(str) }

// IsIMSI checks if a string is valid IMSI
func IsIMSI(str string) bool {
	if !rxIMSI.MatchString(str) {
		return false
	}

	mcc, err := strconv.ParseInt(str[0:3], 10, 32)
	if err != nil {
		return false
	}

	switch mcc {
	case 202, 204, 206, 208, 212, 213, 214, 216, 218, 219:
	case 220, 221, 222, 226, 228, 230, 231, 232, 234, 235:
	case 238, 240, 242, 244, 246, 247, 248, 250, 255, 257:
	case 259, 260, 262, 266, 268, 270, 272, 274, 276, 278:
	case 280, 282, 283, 284, 286, 288, 289, 290, 292, 293:
	case 294, 295, 297, 302, 308, 310, 311, 312, 313, 314:
	case 315, 316, 330, 332, 334, 338, 340, 342, 344, 346:
	case 348, 350, 352, 354, 356, 358, 360, 362, 363, 364:
	case 365, 366, 368, 370, 372, 374, 376, 400, 401, 402:
	case 404, 405, 406, 410, 412, 413, 414, 415, 416, 417:
	case 418, 419, 420, 421, 422, 424, 425, 426, 427, 428:
	case 429, 430, 431, 432, 434, 436, 437, 438, 440, 441:
	case 450, 452, 454, 455, 456, 457, 460, 461, 466, 467:
	case 470, 472, 502, 505, 510, 514, 515, 520, 525, 528:
	case 530, 536, 537, 539, 540, 541, 542, 543, 544, 545:
	case 546, 547, 548, 549, 550, 551, 552, 553, 554, 555:
	case 602, 603, 604, 605, 606, 607, 608, 609, 610, 611:
	case 612, 613, 614, 615, 616, 617, 618, 619, 620, 621:
	case 622, 623, 624, 625, 626, 627, 628, 629, 630, 631:
	case 632, 633, 634, 635, 636, 637, 638, 639, 640, 641:
	case 642, 643, 645, 646, 647, 648, 649, 650, 651, 652:
	case 653, 654, 655, 657, 658, 659, 702, 704, 706, 708:
	case 710, 712, 714, 716, 722, 724, 730, 732, 734, 736:
	case 738, 740, 742, 744, 746, 748, 750, 995:
		return true

	default:
		return false
	}

	return true
}

// IsRsaPublicKey checks if a string is valid public key with provided length.
func IsRsaPublicKey(str string, keylen int) bool {
	pemBytes, err := ioutil.ReadAll(bytes.NewBufferString(str))
	if err != nil {
		return false
	}

	block, _ := pem.Decode(pemBytes)
	if block != nil && block.Type != "PUBLIC KEY" {
		return false
	}

	var der []byte
	if block != nil {
		der = block.Bytes
	} else {
		der, err = base64.StdEncoding.DecodeString(str)
		if err != nil {
			return false
		}
	}

	key, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return false
	}

	pubkey, ok := key.(*rsa.PublicKey)
	if !ok {
		return false
	}

	bitlen := len(pubkey.N.Bytes()) * 8
	return bitlen == int(keylen)
}

// IsRegex checks if a give string is a valid regex with RE2 syntax or not.
func IsRegex(str string) bool {
	if _, err := regexp.Compile(str); err == nil {
		return true
	}
	return false
}

// IsSSN will validate the given string as a U.S. Social Security Number.
func IsSSN(str string) bool {
	if str == "" || len(str) != 11 {
		return false
	}
	return rxSSN.MatchString(str)
}

// IsSemver checks if string is valid semantic version.
func IsSemver(str string) bool { return rxSemver.MatchString(str) }

// IsTime checks if string is valid according to given format.
func IsTime(str string, format string) bool {
	_, err := time.Parse(format, str)
	return err == nil
}

// IsUnixTime checks if string is valid unix timestamp value
func IsUnixTime(str string) bool {
	if _, err := strconv.Atoi(str); err == nil {
		return true
	}
	return false
}

// IsRFC3339 checks if string is valid timestamp value according to RFC3339.
func IsRFC3339(str string) bool { return IsTime(str, time.RFC3339) }

// IsRFC3339WithoutZone checks if string is valid timestamp value according
// to RFC3339 which excludes the timezone.
func IsRFC3339WithoutZone(str string) bool {
	return IsTime(str, rfc3339WithoutZone)
}

// IsE164 checks if string is valid E.164 international phone number.
func IsE164(str string) bool { return rxE164.MatchString(str) }

// Byte to index table for O(1) lookups when unmarshaling.
//
// We use 0xFF as sentinel value for invalid indexes.
var ulidDec = [...]byte{
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x01,
	0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E,
	0x0F, 0x10, 0x11, 0xFF, 0x12, 0x13, 0xFF, 0x14, 0x15, 0xFF,
	0x16, 0x17, 0x18, 0x19, 0x1A, 0xFF, 0x1B, 0x1C, 0x1D, 0x1E,
	0x1F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x0A, 0x0B, 0x0C,
	0x0D, 0x0E, 0x0F, 0x10, 0x11, 0xFF, 0x12, 0x13, 0xFF, 0x14,
	0x15, 0xFF, 0x16, 0x17, 0x18, 0x19, 0x1A, 0xFF, 0x1B, 0x1C,
	0x1D, 0x1E, 0x1F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
}

// EncodedSize is the length of a text encoded ULID.
const ulidEncodedSize = 26

// IsULID checks if the string is a ULID.
//
// Implementation got from:
//   https://github.com/oklog/ulid (Apache-2.0 License)
//
func IsULID(str string) bool {
	// Check if a base32 encoded ULID is the right length.
	if len(str) != ulidEncodedSize {
		return false
	}

	// Check if all the characters in a base32 encoded ULID are part of the
	// expected base32 character set.
	if ulidDec[str[0]] == 0xFF ||
		ulidDec[str[1]] == 0xFF ||
		ulidDec[str[2]] == 0xFF ||
		ulidDec[str[3]] == 0xFF ||
		ulidDec[str[4]] == 0xFF ||
		ulidDec[str[5]] == 0xFF ||
		ulidDec[str[6]] == 0xFF ||
		ulidDec[str[7]] == 0xFF ||
		ulidDec[str[8]] == 0xFF ||
		ulidDec[str[9]] == 0xFF ||
		ulidDec[str[10]] == 0xFF ||
		ulidDec[str[11]] == 0xFF ||
		ulidDec[str[12]] == 0xFF ||
		ulidDec[str[13]] == 0xFF ||
		ulidDec[str[14]] == 0xFF ||
		ulidDec[str[15]] == 0xFF ||
		ulidDec[str[16]] == 0xFF ||
		ulidDec[str[17]] == 0xFF ||
		ulidDec[str[18]] == 0xFF ||
		ulidDec[str[19]] == 0xFF ||
		ulidDec[str[20]] == 0xFF ||
		ulidDec[str[21]] == 0xFF ||
		ulidDec[str[22]] == 0xFF ||
		ulidDec[str[23]] == 0xFF ||
		ulidDec[str[24]] == 0xFF ||
		ulidDec[str[25]] == 0xFF {
		return false
	}

	// Check if the first character in a base32 encoded ULID will overflow. This
	// happens because the base32 representation encodes 130 bits, while the
	// ULID is only 128 bits.
	//
	// See https://github.com/oklog/ulid/issues/9 for details.
	if str[0] > '7' {
		return false
	}
	return true
}
