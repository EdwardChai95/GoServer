package define

const (
	IVNPAY_URL    = "https://apipay.ivnpay.cc" // "https://test-apipay.ivnpay.cc"
	IVNPAY_MCHID  = "102240"
	IVNPAY_SECRET = "e#vib83@Bifbs0b2" // "1111122222333334"

	PAYMENT_RETURNURL = "http://199.115.228.247:12307"

	ORDERSTATUSFAILED  string = "tật nguyền"
	ORDERSTATUSWAITING string = "Chế biến"
	ORDERSTATUSPAID    string = "đã thanh toán"
)

var (
	BLOCKED_IPCODES  = []string{"CN"}
	IPBLOCKEDMESSAGE = "IP blocked"
	// US	United States of America// CN	China// AU	Australia// JP	Japan// TH	Thailand
	// IN	India// MY	Malaysia// KR	Korea (Republic of)// SG	Singapore
	// HK	Hong Kong// TW	Taiwan (Province of China)// KH	Cambodia// PH	Philippines
	// VN	Viet Nam// NO	Norway// ES	Spain// FR	France// NL	Netherlands
	// CZ	Czechia// GB	United Kingdom of Great Britain and Northern Ireland
	// DE	Germany// AT	Austria// CH	Switzerland// BR	Brazil
	// IT	Italy// GR	Greece// PL	Poland// BE	Belgium// IE	Ireland// DK	Denmark
	// PT	Portugal// SE	Sweden// GH	Ghana// TR	Turkey// RU	Russian Federation
	// CM	Cameroon// ZA	South Africa// FI	Finland
	// AE	United Arab Emirates// HU	Hungary	// JO	Jordan// RO	Romania
	// LU	Luxembourg// AR	Argentina	// UG	Uganda// AM	Armenia
	// TZ	Tanzania, United Republic of// BI	Burundi
	// UY	Uruguay// CL	Chile	// BG	Bulgaria// UA	Ukraine
	// EG	Egypt// CA	Canada	// IL	Israel// QA	Qatar	// MD	Moldova (Republic of)// HR	Croatia
	// SC	Seychelles// IQ	Iraq	// LV	Latvia// EE	Estonia// UZ	Uzbekistan// SK	Slovakia// KZ	Kazakhstan
	// GE	Georgia// AL	Albania	// PS	Palestine, State of// SA	Saudi Arabia// CY	Cyprus// MT	Malta
	// LT	Lithuania// CR	Costa Rica// IR	Iran (Islamic Republic of)// BH	Bahrain
	// MX	Mexico// CO	Colombia// SY	Syrian Arab Republic// LB	Lebanon
	// AZ	Azerbaijan// ZW	Zimbabwe// ZM	Zambia// OM	Oman
	// RS	Serbia// IS	Iceland// SI	Slovenia// MK	North Macedonia
	// LI	Liechtenstein// JE	Jersey// BA	Bosnia and Herzegovina
	// PE	Peru// KG	Kyrgyzstan// IM	Isle of Man// GG	Guernsey// GI	Gibraltar
	// LY	Libya// YE	Yemen// BY	Belarus// RE	Reunion
	// JM	Jamaica  	// GP	Guadeloupe  	// MQ	Martinique  	// KW	Kuwait
	// LK	Sri Lanka  	// SZ	Eswatini  	// CD	Congo (Democratic Republic of the)
	// BT	Bhutan  	// BN	Brunei Darussalam  	// PM	Saint Pierre and Miquelon
	// PA	Panama  	// LA	Lao People's Democratic Republic  	// GU	Guam
	// MP	Northern Mariana Islands  	// DO	Dominican Republic  	// ID	Indonesia
	// NG	Nigeria  	// NZ	New Zealand  	// EC	Ecuador
	// VE	Venezuela (Bolivarian Republic of)  	// PR	Puerto Rico
	// BO	Bolivia (Plurinational State of)  	// VI	Virgin Islands (U.S.)
	// BD	Bangladesh  	// PK	Pakistan  	// PG	Papua New Guinea
	// TL	Timor-Leste  	// SB	Solomon Islands  	// VU	Vanuatu
	// FJ	Fiji  	// CK	Cook Islands  	// TO	Tonga  	// NP	Nepal  	// KE	Kenya
	// MO	Macao  	// TT	Trinidad and Tobago  	// LS	Lesotho  	// MA	Morocco
	// VG	Virgin Islands (British)  	// KN	Saint Kitts and Nevis  	// AG	Antigua and Barbuda
	// VC	Saint Vincent and The Grenadines  	// KY	Cayman Islands  	// LC	Saint Lucia
	// MM	Myanmar  	// GD	Grenada  	// CW	Curacao  	// BB	Barbados
	// BS	Bahamas  	// PY	Paraguay  	// GT	Guatemala  	// UM	United States Minor Outlying Islands
	// DM	Dominica  	// TM	Turkmenistan  	// TK	Tokelau  	// MV	Maldives
	// AF	Afghanistan  	// NC	New Caledonia  	// MN	Mongolia  	// WF	Wallis and Futuna
	// SM	San Marino  	// ME	Montenegro  	// SV	El Salvador  	// AD	Andorra
	// MC	Monaco  	// GL	Greenland  	// BZ	Belize  	// TJ	Tajikistan
	// FO	Faroe Islands  	// HT	Haiti  	// MF	Saint Martin (French Part)
	// LR	Liberia  	// MU	Mauritius  	// BW	Botswana  	// TN	Tunisia  	// MG	Madagascar
	// AO	Angola  	// NA	Namibia  	// CI	Cote D'ivoire  	// SD	Sudan
	// MW	Malawi  	// GA	Gabon  	// ML	Mali  	// BJ	Benin  	// TD	Chad
	// CV	Cabo Verde  	// RW	Rwanda  	// CG	Congo  	// MZ	Mozambique
	// GM	Gambia  	// GN	Guinea  	// BF	Burkina Faso  	// SO	Somalia  	// SL	Sierra Leone
	// NE	Niger  	// CF	Central African Republic  	// TG	Togo  	// SS	South Sudan
	// GQ	Equatorial Guinea  	// SN	Senegal  	// DZ	Algeria  	// AS	American Samoa
	// MR	Mauritania  	// DJ	Djibouti  	// KM	Comoros  	// IO	British Indian Ocean Territory
	// YT	Mayotte  	// NR	Nauru  	// WS	Samoa  	// FM	Micronesia (Federated States of)
	// PF	French Polynesia  	// HN	Honduras  	// NI	Nicaragua  	// BM	Bermuda
	// GF	French Guiana  	// NU	Niue  	// TV	Tuvalu  	// PW	Palau  	// MH	Marshall Islands
	// KI	Kiribati  	// KP	Korea (Democratic People's Republic of)  	// AW	Aruba  	// CU	Cuba
	// SR	Suriname  	// GY	Guyana  	// VA	Holy See  	// ST	Sao Tome and Principe
	// ET	Ethiopia  	// ER	Eritrea  	// GW	Guinea-Bissau  	// FK	Falkland Islands (Malvinas)
	// BL	Saint Barthelemy  	// AI	Anguilla  	// TC	Turks and Caicos Islands  	// SX	Sint Maarten (Dutch Part)
	// AX	Aland Islands  	// AQ	Antarctica  	// NF	Norfolk Island
	// BQ	Bonaire, Sint Eustatius and Saba  	// MS	Montserrat
	// GS	South Georgia and The South Sandwich Islands  	// SJ	Svalbard and Jan Mayen
	// SH	Saint Helena, Ascension and Tristan Da Cunha

	GopokerSigningKey = []byte("haveitoldyouiloveyoulately...")

	EthWalletPrivateSecret = "pass5871"
)