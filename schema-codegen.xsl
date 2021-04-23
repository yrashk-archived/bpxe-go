<?xml version="1.0" encoding="UTF-8"?>
<xsl:stylesheet xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
    xmlns:xs="http://www.w3.org/2001/XMLSchema"
    xmlns:local="http://bpxe.org"
    exclude-result-prefixes="xs"
    version="3.0">
    
    <xsl:variable name="schema"
        select="(/ | document(/xs:schema/xs:include/@schemaLocation))/xs:schema"/>
    <xsl:variable name="elements" select="$schema//xs:element"/>
    
    <xsl:strip-space elements="*"/>
    
    <xsl:template match="/">
        
        <xsl:result-document method="text" href="schema_generated.go">
            <xsl:text>
                package bpmn
                // This file is generated from BPMN 2.0 schema using `make generate`
                // DO NOT EDIT                
                import ( "math/big" )
                
            </xsl:text>
            
            <xsl:for-each select="$schema/xs:simpleType[@name]">
                <xsl:text xml:space="preserve">type </xsl:text>
                <xsl:value-of select="local:type(.)"/>
                <xsl:text xml:space="preserve"> </xsl:text>
                <xsl:choose>
                    <xsl:when test="exists(./xs:restriction)">
                        <xsl:value-of select="local:type(./xs:restriction/@base)"/>
                    </xsl:when>
                    <xsl:when test="exists(./xs:union[@memberTypes])">
                        <xsl:value-of select="local:type(./xs:union/@memberTypes)"/>
                    </xsl:when>
                </xsl:choose>
                <xsl:text xml:space="preserve">
                </xsl:text>
            </xsl:for-each>
            
            <xsl:for-each select="$schema/xs:complexType[@name]">
                <xsl:call-template name="type">
                    <xsl:with-param name="type" select="." />
                </xsl:call-template>
            </xsl:for-each>
            
        </xsl:result-document>
        
        <xsl:result-document method="text" href="schema_generated_test.go">
            <xsl:text>
                // This content is generated from BPMN 2.0 schema using `make generate`
                // DO NOT EDIT
                package bpmn
                
                import "testing"
            </xsl:text>
            
            <xsl:for-each select="$schema/xs:complexType[@name]">
                <xsl:call-template name="type-test">
                    <xsl:with-param name="type" select="." />
                </xsl:call-template>
            </xsl:for-each>
        </xsl:result-document>
    </xsl:template>
    
    <xsl:template name="type">
        <xsl:param name="type" required="yes"/>
        <!-- Type -->
        
        
        <xsl:text>type </xsl:text>
        <xsl:value-of select="local:struct-case($type/@name)"/>
        <xsl:text> struct {  </xsl:text>
        <xsl:if test="exists(./xs:complexContent/xs:extension[@base])">
            <xsl:value-of select="local:struct-case(./xs:complexContent/xs:extension/@base)"/>
            <xsl:text xml:space="preserve">
            </xsl:text>
        </xsl:if>
        <xsl:for-each select=".//xs:attribute">
            <xsl:value-of select="local:field-name(.)"/>
            <xsl:text xml:space="preserve"> </xsl:text>
            <xsl:value-of select="local:field-type(.)"/>
            <xsl:text xml:space="preserve"> `xml:"</xsl:text>
            <xsl:value-of select="./@name"/>
            <xsl:text xml:space="preserve">,attr"`
            </xsl:text>
        </xsl:for-each>
        <xsl:for-each select="local:specific-elements(.)">
            <xsl:choose>
                <xsl:when test="exists(./@name)">
                    <xsl:value-of select="local:field-name(.)"/>
                    <xsl:text xml:space="preserve"> </xsl:text>
                    <xsl:value-of select="local:field-type(.)"/>
                    <xsl:text xml:space="preserve"> `xml:"http://www.omg.org/spec/BPMN/20100524/MODEL </xsl:text>
                    <xsl:value-of select="./@name"/>
                    <xsl:text xml:space="preserve">"`
                    </xsl:text>
                </xsl:when>
                <xsl:when test="local:is-a-ref(.)">
                    <xsl:value-of select="local:field-name(.)"/>
                    <xsl:text xml:space="preserve"> </xsl:text>
                    <xsl:value-of select="local:ref-type(.)"/>
                    <xsl:text xml:space="preserve"> `xml:"http://www.omg.org/spec/BPMN/20100524/MODEL </xsl:text>
                    <xsl:value-of select="./@ref"/>
                    <xsl:text xml:space="preserve">"`
                    </xsl:text>
                </xsl:when>
                <xsl:otherwise/>
            </xsl:choose>
        </xsl:for-each>
        <xsl:text>TextPayloadField string `xml:",chardata"`</xsl:text>
        <xsl:text xml:space="preserve"> }
        </xsl:text>
        <!-- Constructor -->
        
        <xsl:variable name="quot">"</xsl:variable>
        
        <!-- defaults -->
        
        <xsl:for-each select=".//xs:attribute[exists(@default)]">
            <xsl:if test="not(contains(./@default, '#'))">
                <xsl:text>var default</xsl:text>
                <xsl:value-of select="local:struct-case($type/@name)"/>
                <xsl:value-of select="local:field-name(.)"/>
                <xsl:text xml:spce="preserve"> </xsl:text>
                <xsl:value-of select="local:type(./@type)"/>
                <xsl:text>=</xsl:text>
                <xsl:choose>
                    <xsl:when test="./@type = 'xsd:boolean'">
                        <xsl:value-of select="./@default"/>
                    </xsl:when>
                    <xsl:when test="./@type = 'xsd:int'">
                        <xsl:value-of select="./@default"/>
                    </xsl:when>
                    <xsl:when test="./@type = 'xsd:string' or ./@type = 'xsd:anyURI'">
                        <xsl:value-of select="concat($quot, ./@default, $quot)"/>
                    </xsl:when>
                    <xsl:when test="./@type = 'xsd:integer'">
                        <xsl:value-of select="concat('*big.NewInt(', ./@default, ')')"/>
                    </xsl:when>
                    <xsl:when test="not(contains(./@type, 'xsd:'))">
                        <xsl:value-of select="concat($quot, ./@default, $quot)"/>
                    </xsl:when>
                    <xsl:otherwise>
                        <xsl:value-of select="./@default"/>
                    </xsl:otherwise>
                </xsl:choose>
                <xsl:text xml:space="preserve">
                </xsl:text>
            </xsl:if>
        </xsl:for-each>
        
        <xsl:text xml:space="preserve">func Default</xsl:text>
        <xsl:value-of select="local:struct-case($type/@name)"/>
        <xsl:text>()</xsl:text>
        <xsl:value-of select="local:struct-case($type/@name)"/>
        <xsl:text>{</xsl:text>
        <xsl:text xml:space="preserve">return </xsl:text>
        <xsl:value-of select="local:struct-case($type/@name)"/>
        <xsl:text xml:space="preserve">{
        </xsl:text>
        
        <xsl:if test="exists(./xs:complexContent/xs:extension[@base])">
            <xsl:value-of select="local:struct-case(./xs:complexContent/xs:extension/@base)"/>
            <xsl:text>: Default</xsl:text>
            <xsl:value-of select="local:struct-case(./xs:complexContent/xs:extension/@base)"/>
            <xsl:text xml:space="preserve">(),
            </xsl:text>
        </xsl:if>
        
        <xsl:for-each select=".//xs:attribute[exists(@default)]">
            <xsl:if test="not(contains(./@default, '#'))">
                <xsl:value-of select="local:field-name(.)"/>
                <xsl:text>:</xsl:text>
                <xsl:if test="local:is-optional(.)">
                    <xsl:text>&amp;</xsl:text>
                </xsl:if>
                <xsl:choose>
                    <xsl:when test="./@type = 'xsd:string' or ./@type = 'xsd:anyURI'">
                        <xsl:text>default</xsl:text>
                        <xsl:value-of select="local:struct-case($type/@name)"/>
                        <xsl:value-of select="local:field-name(.)"/>
                    </xsl:when>
                    <xsl:otherwise>
                        <xsl:text>default</xsl:text>
                        <xsl:value-of select="local:struct-case($type/@name)"/>
                        <xsl:value-of select="local:field-name(.)"/>
                    </xsl:otherwise>
                </xsl:choose>
                <xsl:text xml:space="preserve">,
                </xsl:text>
            </xsl:if>
        </xsl:for-each>
        <xsl:text>}</xsl:text>
        <xsl:text xml:space="preserve">}
        </xsl:text>
        <!-- Interface -->
        
        <xsl:text>type </xsl:text>
        <xsl:value-of select="local:struct-case($type/@name)"/>
        <xsl:text xml:space="preserve">Interface interface {
            Element
        </xsl:text>
        
        <xsl:text>
            TextPayload() *string
        </xsl:text>
        
        <xsl:if test="exists(./xs:complexContent/xs:extension[@base])">
            <xsl:value-of select="local:struct-case(./xs:complexContent/xs:extension/@base)"/>
            <xsl:text xml:space="preserve">Interface
            </xsl:text>
        </xsl:if>
        <!-- Getters -->
        <xsl:for-each select=".//xs:attribute">
            <xsl:value-of select="local:struct-case(./@name)"/>
            <xsl:text xml:space="preserve">() (result </xsl:text>
            <xsl:value-of select="local:returning-type(.)"/>
            <xsl:if test="local:is-optional(.)">
                <xsl:text xml:space="preserve">, present bool</xsl:text>
            </xsl:if>
            <xsl:text xml:space="preserve">)
            </xsl:text>
        </xsl:for-each>
        <xsl:for-each select="local:specific-elements(.)">
            <xsl:choose>
                <xsl:when test="exists(./@name)">
                    <xsl:value-of select="local:field-method(.)"/>
                    <xsl:text xml:space="preserve">() (result </xsl:text>
                    <xsl:value-of select="local:returning-type(.)"/>
                    <xsl:if test="local:is-optional(.)">
                        <xsl:text xml:space="preserve">, present bool</xsl:text>
                    </xsl:if>
                    <xsl:text xml:space="preserve">)
                    </xsl:text>
                </xsl:when>
                <xsl:when test="local:is-a-ref(.)">
                    <xsl:value-of select="local:field-method(.)"/>
                    <xsl:text xml:space="preserve">() (result </xsl:text>
                    <xsl:value-of select="local:returning-ref-type(.)"/>
                    <xsl:if test="local:is-optional(.)">
                        <xsl:text xml:space="preserve">, present bool</xsl:text>
                    </xsl:if>
                    <xsl:text xml:space="preserve">)
                    </xsl:text>
                </xsl:when>
                <xsl:otherwise/>
            </xsl:choose>
        </xsl:for-each>
        <!-- Setters -->
        <xsl:for-each select=".//xs:attribute">
            <xsl:text>Set</xsl:text>
            <xsl:value-of select="local:struct-case(./@name)"/>
            <xsl:text xml:space="preserve">(value </xsl:text>
            <xsl:value-of select="local:type(./@type)"/>
            <xsl:text>) </xsl:text>
            <xsl:text xml:space="preserve">
            </xsl:text>
        </xsl:for-each>
        <xsl:for-each select="local:specific-elements(.)">
            <xsl:choose>
                <xsl:when test="exists(./@name)">
                    <xsl:text>Set</xsl:text>
                    <xsl:value-of select="local:field-method(.)"/>
                    <xsl:text xml:space="preserve">(value </xsl:text>
                    <xsl:value-of select="local:param-type(.)"/>
                    <xsl:text>) </xsl:text>
                    <xsl:text xml:space="preserve">
                    </xsl:text>
                </xsl:when>
                <xsl:when test="local:is-a-ref(.)">
                    <xsl:text>Set</xsl:text>
                    <xsl:value-of select="local:field-method(.)"/>
                    <xsl:text xml:space="preserve">(value </xsl:text>
                    <xsl:value-of select="local:ref-type(.)"/>
                    <xsl:text>) </xsl:text>
                    <xsl:text xml:space="preserve">
                    </xsl:text>
                </xsl:when>
                <xsl:otherwise/>
            </xsl:choose>
        </xsl:for-each>
        <xsl:text xml:space="preserve"> }
        </xsl:text>
        <!-- Interface implementation -->
        <xsl:text xml:space="preserve">
            func (t *</xsl:text><xsl:value-of select="local:struct-case($type/@name)"/>
        <xsl:text xml:space="preserve">) TextPayload() *string {
            return &amp;t.TextPayloadField
         }
        </xsl:text>
        
        <xsl:text xml:space="preserve">func (t *</xsl:text><xsl:value-of select="local:struct-case($type/@name)"/>
        <xsl:text xml:space="preserve">) FindBy(f ElementPredicate) (result Element, found bool) {
            if f(t) {
            result = t
            found = true
            return
            }
        </xsl:text>
        <xsl:if test="exists(./xs:complexContent/xs:extension[@base])">
            <xsl:text>if result, found = t.</xsl:text><xsl:value-of select="local:struct-case(./xs:complexContent/xs:extension/@base)"/>
            <xsl:text xml:space="preserve">.FindBy(f); found {
                return
            }
            </xsl:text>
        </xsl:if>
        <xsl:for-each select="local:specific-elements(.)">
            <xsl:variable name="ref" select="./@ref"/>
            <xsl:choose>
                <xsl:when test="contains($ref, ':') or contains(./@type, ':')">
                    <!-- other schemas are of no interest -->
                </xsl:when>
                <xsl:when test="exists($schema/xs:simpleType[@name = $ref]) or exists($schema/xs:simpleType[@name = $type])">
                    <!-- don't search simple types -->
                </xsl:when>
                <xsl:when test="./@maxOccurs = 'unbounded'">
                    <xsl:text>
                        for _, e := range t.</xsl:text><xsl:value-of select="local:field-name(.)"/><xsl:text xml:space="preserve">{
                        if result, found = e.FindBy(f); found {
                               return
                         }
                         }
                     </xsl:text>
                </xsl:when>
                <xsl:when test="./@minOccurs = '0' and ./@maxOccurs = '1'">
                    <xsl:text>
                        if value := t.</xsl:text><xsl:value-of select="local:field-name(.)"/><xsl:text xml:space="preserve">; value != nil {
                        if result, found = value.FindBy(f); found {
                               return
                         }
                         }
                        </xsl:text>
                </xsl:when>
                <xsl:otherwise>
                    <xsl:text>
                        if result, found = t.</xsl:text><xsl:value-of select="local:field-name(.)"/><xsl:text xml:space="preserve">.FindBy(f); found {
                               return
                         }
                        </xsl:text>
                </xsl:otherwise>
            </xsl:choose>
            
        </xsl:for-each>
        <xsl:text xml:space="preserve">
          return    
        }
        </xsl:text>
        <!-- attributes -->
        <xsl:for-each select=".//xs:attribute">
            <!-- Getter -->
            <xsl:text xml:space="preserve">func (t *</xsl:text>
            <xsl:value-of select="local:struct-case($type/@name)"/>
            <xsl:text xml:space="preserve">) </xsl:text>
            <xsl:value-of select="local:field-method(.)"/>
            <xsl:text xml:space="preserve">() (result </xsl:text>
            <xsl:value-of select="local:returning-type(.)"/>
            <xsl:if test="local:is-optional(.)">
                <xsl:text xml:space="preserve">, present bool</xsl:text>
            </xsl:if>
            <xsl:text>){
            </xsl:text>
            <xsl:if test="local:is-optional(.)">
                <xsl:text>if t.</xsl:text>
                <xsl:value-of select="local:field-name(.)"/>
                <xsl:text> != nil {
                    present = true
                    }
                </xsl:text>
            </xsl:if>
            <xsl:text>    result = </xsl:text>
            <xsl:value-of select="local:returning(.)"/>
            <xsl:text>t.</xsl:text>
            <xsl:value-of select="local:field-name(.)"/>
            <xsl:text xml:space="preserve">
                return
                }
            </xsl:text>
            <!-- Setter -->
            <xsl:text xml:space="preserve">func (t *</xsl:text>
            <xsl:value-of select="local:struct-case($type/@name)"/>
            <xsl:text xml:space="preserve">) </xsl:text>
            <xsl:text>Set</xsl:text>
            <xsl:value-of select="local:field-method(.)"/>
            <xsl:text xml:space="preserve">(value </xsl:text>
            <xsl:value-of select="local:type(./@type)"/>
            <xsl:text xml:space="preserve">) </xsl:text>
            <xsl:text xml:space="preserve">{
            </xsl:text>
            <xsl:text>t.</xsl:text>
            <xsl:value-of select="local:field-name(.)"/>
            <xsl:text> = </xsl:text>
            <xsl:if test="local:is-optional(.)">
                <xsl:text>&amp;</xsl:text>
            </xsl:if>
            <xsl:text>value</xsl:text>
            <xsl:text xml:space="preserve">
                }
            </xsl:text>
        </xsl:for-each>
        <!-- elements -->
        <xsl:for-each select="local:specific-elements(.)">
            <!-- Getter -->
            <xsl:choose>
                <xsl:when test="exists(./@name)">
                    <xsl:text xml:space="preserve">func (t *</xsl:text>
                    <xsl:value-of select="local:struct-case($type/@name)"/>
                    <xsl:text xml:space="preserve">) </xsl:text>
                    <xsl:value-of select="local:field-method(.)"/>
                    <xsl:text xml:space="preserve">() (result </xsl:text>
                    <xsl:value-of select="local:returning-type(.)"/>
                    <xsl:if test="local:is-optional(.)">
                        <xsl:text xml:space="preserve">, present bool</xsl:text>
                    </xsl:if>
                    <xsl:text>) {</xsl:text>
                    <xsl:if test="local:is-optional(.)">
                        <xsl:text>if t.</xsl:text>
                        <xsl:value-of select="local:field-name(.)"/>
                        <xsl:text> != nil {
                            present = true
                            }
                        </xsl:text>
                    </xsl:if>
                    <xsl:text>    result = </xsl:text>
                    <xsl:value-of select="local:returning(.)"/>
                    <xsl:text>t.</xsl:text>
                    <xsl:value-of select="local:field-name(.)"/>
                    <xsl:text xml:space="preserve">
                        return
                        }
                    </xsl:text>
                </xsl:when>
                <xsl:when test="local:is-a-ref(.)">
                    <xsl:text xml:space="preserve">func (t *</xsl:text>
                    <xsl:value-of select="local:struct-case($type/@name)"/>
                    <xsl:text xml:space="preserve">) </xsl:text>
                    <xsl:value-of select="local:field-method(.)"/>
                    <xsl:text xml:space="preserve">() (result </xsl:text>
                    <xsl:value-of select="local:returning-ref-type(.)"/>
                    <xsl:if test="local:is-optional(.)">
                        <xsl:text xml:space="preserve">, present bool</xsl:text>
                    </xsl:if>
                    <xsl:text>) {</xsl:text>
                    <xsl:if test="local:is-optional(.)">
                        <xsl:text>if t.</xsl:text>
                        <xsl:value-of select="local:field-name(.)"/>
                        <xsl:text> != nil {
                            present = true
                            }
                        </xsl:text>
                    </xsl:if>
                    <xsl:text>    result = </xsl:text>
                    <xsl:value-of select="local:returning(.)"/>
                    <xsl:text>t.</xsl:text>
                    <xsl:value-of select="local:field-name(.)"/>
                    <xsl:text xml:space="preserve">
                        return
                        }
                    </xsl:text>
                </xsl:when>
                <xsl:otherwise/>
            </xsl:choose>
            <!-- Setter -->
            <xsl:choose>
                <xsl:when test="exists(./@name)">
                    <xsl:text xml:space="preserve">func (t *</xsl:text>
                    <xsl:value-of select="local:struct-case($type/@name)"/>
                    <xsl:text>) Set</xsl:text>
                    <xsl:value-of select="local:field-method(.)"/>
                    <xsl:text xml:space="preserve">(value </xsl:text>
                    <xsl:value-of select="local:param-type(.)"/>
                    <xsl:text>) </xsl:text>
                    <xsl:text xml:space="preserve">{
                    </xsl:text>
                    <xsl:text>t.</xsl:text>
                    <xsl:value-of select="local:field-name(.)"/>
                    <xsl:text> = </xsl:text>
                    <xsl:if test="local:is-optional(.)">
                        <xsl:text>&amp;</xsl:text>
                    </xsl:if>
                    <xsl:text>value</xsl:text>
                    <xsl:text xml:space="preserve">
                        }
                    </xsl:text>
                </xsl:when>
                <xsl:when test="local:is-a-ref(.)">
                    <xsl:text xml:space="preserve">func (t *</xsl:text>
                    <xsl:value-of select="local:struct-case($type/@name)"/>
                    <xsl:text>) Set</xsl:text>
                    <xsl:value-of select="local:field-method(.)"/>
                    <xsl:text xml:space="preserve">(value </xsl:text>
                    <xsl:value-of select="local:ref-type(.)"/>
                    <xsl:text>) </xsl:text>
                    <xsl:text xml:space="preserve">{
                    </xsl:text>
                    <xsl:text>t.</xsl:text>
                    <xsl:value-of select="local:field-name(.)"/>
                    <xsl:text> = </xsl:text>
                    <!--<xsl:if test="local:is-optional(.)">
                        <xsl:text>&amp;</xsl:text>
                        </xsl:if>-->
                    <xsl:text>value</xsl:text>
                    <xsl:text xml:space="preserve">
                        }
                    </xsl:text>
                </xsl:when>
                <xsl:otherwise/>
            </xsl:choose>
        </xsl:for-each>
    </xsl:template>
    
    <xsl:template name="type-test">
        <xsl:param name="type" required="yes"/>
        
        <xsl:text xml:space="preserve">func Test</xsl:text><xsl:value-of select="local:struct-case($type/@name)"/><xsl:text xml:space="preserve">Interface(t *testing.T) {
            // won't compile if the interaces are not implemented
            var _ </xsl:text>
        <xsl:value-of select="local:struct-case($type/@name)"/>
        <xsl:text xml:space="preserve">Interface = &amp;</xsl:text>
        <xsl:value-of select="local:struct-case($type/@name)"/>
        <xsl:text xml:space="preserve"> {}
        </xsl:text>
        <xsl:text xml:space="preserve">
            }
        </xsl:text>
    </xsl:template>
    
    <xsl:function name="local:struct-case">
        <xsl:param name="string"/>
        <xsl:choose>
            <xsl:when test="matches($string, '^t[A-Z][A-Za-z0-9_]+')">
                <xsl:sequence select="local:struct-case(substring($string, 2, string-length($string) - 1))"/>
            </xsl:when>
            <xsl:otherwise>
                <xsl:sequence select="
                    string-join(for $s in tokenize($string, '\W+')
                    return
                    concat(upper-case(substring($s, 1, 1)), substring($s, 2)), '')"/>
            </xsl:otherwise>
        </xsl:choose>
    </xsl:function>
    
    <xsl:function name="local:field-name">
        <xsl:param name="el"/>
        <xsl:variable name="string" select="if (exists($el/@ref)) then $el/@ref else $el/@name"/>
        <xsl:choose>
            <xsl:when test="matches($string, '^t[A-Z][A-Za-z0-9_]+')">
                <xsl:sequence select="local:field-name(substring($string, 2, string-length($string) - 1))"/>
            </xsl:when>
            <xsl:otherwise>
                <xsl:sequence select="
                    string-join(for $s in tokenize($string, '\W+')
                    return
                    concat(upper-case(substring($s, 1, 1)), substring($s, 2),'Field'), '')"/>
            </xsl:otherwise>
        </xsl:choose>
    </xsl:function>
    
    <xsl:function name="local:pluralize">
        <xsl:param name="word"/>
        <xsl:param name="plural" as="xs:boolean"/>
        <xsl:choose>
            <xsl:when test="$plural">
                <xsl:choose>
                    <xsl:when test="ends-with($word, 'ss') or ends-with($word, 'x') or ends-with($word, 'ch') or ends-with($word, 'sh')"><xsl:sequence select="concat($word, 'es')"/></xsl:when>
                    <xsl:when test="ends-with($word, 's')"><xsl:sequence select="concat($word, 'es')"/></xsl:when>
                    <xsl:when test="ends-with($word, 'ey') or ends-with($word, 'ay') or ends-with($word, 'oy')"><xsl:sequence select="concat($word, 's')"/></xsl:when>
                    <xsl:when test="ends-with($word, 'y')"><xsl:sequence select="concat(substring($word,0, string-length($word)), 'ies')"/></xsl:when>
                    <xsl:otherwise>
                        <xsl:sequence select="concat($word, 's')"/>
                    </xsl:otherwise>
                </xsl:choose>
            </xsl:when>
            <xsl:otherwise>
                <xsl:sequence select="$word"/>
            </xsl:otherwise>
        </xsl:choose>
    </xsl:function>
    
    <xsl:function name="local:field-method">
        <xsl:param name="el"/>
        <xsl:variable name="plurality" select="if ($el/@maxOccurs = 'unbounded') then true() else false()"/>
        <xsl:variable name="string" select="if (exists($el/@ref)) then $el/@ref else $el/@name"/>
        <xsl:choose>
            <xsl:when test="matches($string, '^t[A-Z][A-Za-z0-9_]+')">
                <xsl:sequence select="local:field-name(substring($string, 2, string-length($string) - 1))"/>
            </xsl:when>
            <xsl:otherwise>
                <xsl:sequence select="
                    string-join(for $s in tokenize($string, '\W+')
                    return
                    local:pluralize(concat(upper-case(substring($s, 1, 1)), substring($s, 2)), $plurality), '')"/>
            </xsl:otherwise>
        </xsl:choose>
    </xsl:function>
    
    
    <xsl:function name="local:type">
        <xsl:param name="el"/>
        <xsl:variable name="name" select="if (exists($el/@name)) then $el/@name else $el"/>
        <xsl:choose>
            <xsl:when test="$el/@maxOccurs = 'unbounded'">
                <xsl:sequence select="concat('[]', local:type($name))"/>
            </xsl:when>
            <xsl:when test="$name = 'tExpression'"><xsl:text>AnExpression</xsl:text></xsl:when>
            <xsl:when test="$name = 'xsd:boolean'"><xsl:text>bool</xsl:text></xsl:when>
            <xsl:when test="$name = 'xsd:string'"><xsl:text>string</xsl:text></xsl:when>
            <xsl:when test="$name = 'xsd:QName'"><xsl:text>QName</xsl:text></xsl:when>
            <xsl:when test="$name = 'xsd:integer'"><xsl:text>big.Int</xsl:text></xsl:when>
            <xsl:when test="$name = 'xsd:int'"><xsl:text>int32</xsl:text></xsl:when>
            <xsl:when test="$name = 'xsd:IDREF'"><xsl:text>IdRef</xsl:text></xsl:when>
            <xsl:when test="$name = 'xsd:ID'"><xsl:text>Id</xsl:text></xsl:when>
            <xsl:when test="$name = 'xsd:anyURI'"><xsl:text>AnyURI</xsl:text></xsl:when>
            <xsl:otherwise><xsl:sequence select="local:struct-case($name)"/></xsl:otherwise>
        </xsl:choose>
    </xsl:function>
    
    <xsl:function name="local:param-type">
        <xsl:param name="el"/>
        <xsl:variable name="name" select="if (exists($el/@type)) then $el/@type else $el"/>
        <xsl:choose>
            <xsl:when test="$el/@maxOccurs = 'unbounded'">
                <xsl:sequence select="concat('[]', local:type($name))"/>
            </xsl:when>
            <xsl:otherwise><xsl:sequence select="local:type($name)"/></xsl:otherwise>
        </xsl:choose>
    </xsl:function>
    
    <xsl:function name="local:returning-type-internal">
        <xsl:param name="el"/>
        <xsl:variable name="ref" select="$el/@type"/>
        <xsl:choose>
            <xsl:when test="$el/@maxOccurs = 'unbounded'">
                <xsl:sequence select="concat('*[]', local:type($ref))"/>
            </xsl:when>
            <xsl:when test="$ref= 'tExpression'"><xsl:text>*AnExpression</xsl:text></xsl:when>
            <xsl:when test="$ref = 'xsd:boolean'"><xsl:text>bool</xsl:text></xsl:when>
            <xsl:when test="$ref = 'xsd:int'"><xsl:text>int32</xsl:text></xsl:when>
            <xsl:when test="$ref = 'xsd:string'"><xsl:text>*string</xsl:text></xsl:when>
            <xsl:when test="$ref = 'xsd:QName'"><xsl:text>*QName</xsl:text></xsl:when>
            <xsl:when test="$ref = 'xsd:integer'"><xsl:text>*big.Int</xsl:text></xsl:when>
            <xsl:when test="$ref = 'xsd:IDREF'"><xsl:text>*IdRef</xsl:text></xsl:when>
            <xsl:when test="$ref = 'xsd:ID'"><xsl:text>*Id</xsl:text></xsl:when>
            <xsl:when test="$ref = 'xsd:anyURI'"><xsl:text>*AnyURI</xsl:text></xsl:when>
            <xsl:otherwise><xsl:sequence select="concat('*', local:struct-case($ref))"/></xsl:otherwise>
        </xsl:choose>
    </xsl:function>
    
    
    <xsl:function name="local:returning-type">
        <xsl:param name="el"/>
        <xsl:variable name="result" select="local:returning-type-internal($el)"/>
        <xsl:choose>
            <xsl:when test="local:is-optional($el) and not(contains($result,'*'))">
                <xsl:sequence select="concat('*', $result)"/>
            </xsl:when>
            <xsl:otherwise><xsl:sequence select="$result"/></xsl:otherwise>
        </xsl:choose>
    </xsl:function>
    
    <xsl:function name="local:returning">
        <xsl:param name="el"/>
        <xsl:choose>
            <xsl:when test="local:is-optional($el)"></xsl:when>
            <xsl:when test="$el/@type = 'xsd:boolean'"></xsl:when>
            <xsl:when test="$el/@type = 'xsd:int'"></xsl:when>
            <xsl:when test="$el/@type = 'xsd:int32'"></xsl:when>
            <xsl:otherwise><xsl:text>&amp;</xsl:text></xsl:otherwise>
        </xsl:choose>
    </xsl:function>
    
    
    <xsl:function name="local:ref-type">
        <xsl:param name="el"/>
        <xsl:variable name="ref" select="$el/@ref"/>
        <xsl:variable name="name" select="$schema/xs:element[@name = $ref]/@type"/>
        <xsl:choose>
            <xsl:when test="$el/@maxOccurs = 'unbounded'">
                <xsl:sequence select="concat('[]',local:struct-case($name))"/>
            </xsl:when>
            <xsl:when test="local:is-optional($el)">
                <xsl:sequence select="concat('*',local:struct-case($name))"/>
            </xsl:when>
            <xsl:otherwise><xsl:sequence select="local:struct-case($name)"/></xsl:otherwise>
        </xsl:choose>
    </xsl:function>
    
    <xsl:function name="local:returning-ref-type">
        <xsl:param name="ref"/>
        <xsl:variable name="refType" select="local:ref-type($ref)"/>
        <xsl:choose>
            <xsl:when test="contains($refType,'*')"><xsl:sequence select="$refType"/></xsl:when>
            <xsl:otherwise><xsl:sequence select="concat('*', local:ref-type($ref))"/></xsl:otherwise>
        </xsl:choose>
    </xsl:function>
    
    <xsl:function name="local:field-type">
        <xsl:param name="el"/>
        <xsl:choose>
            <!-- Account for recursion -->
            <xsl:when test="local:type($el/@type) = 'LaneSet'">*LaneSet</xsl:when>
            <xsl:when test="$el/@maxOccurs = 'unbounded'">
                <xsl:sequence select="concat('[]',local:type($el/@type))"/>
            </xsl:when>
            <xsl:when test="local:is-optional($el)">
                <xsl:sequence select="concat('*',local:type($el/@type))"/>
            </xsl:when>
            <xsl:otherwise>
                <xsl:sequence select="local:type($el/@type)"/>
            </xsl:otherwise>
        </xsl:choose>
    </xsl:function>
    
    <xsl:function name="local:is-a-ref">
        <xsl:param name="el"/>
        <xsl:sequence select="exists($el/@ref) and not(contains($el/@ref, ':'))"/>
    </xsl:function>
    
    <xsl:function name="local:is-optional">
        <xsl:param name="el"/>
        <xsl:choose>
            <xsl:when test="$el/@use = 'optional'"><xsl:sequence select="true()"/></xsl:when>
            <xsl:when test="$el/@minOccurs = '0' and $el/@maxOccurs = '1'"><xsl:sequence select="true()"/></xsl:when>
            <xsl:otherwise><xsl:sequence select="false()"/></xsl:otherwise>
        </xsl:choose>
    </xsl:function>
    
    <xsl:function name="local:specific-elements">
        <xsl:param name="el"/>
        <xsl:variable name="elements" select="copy-of($el//xs:element)"/>
        <xsl:for-each select="$elements">
            <xsl:variable name="element" select="."/>
            <xsl:variable name="name" select="./@ref"/>
            <xsl:variable name="ref" select="$schema/xs:element[@name = $name]"/>
            <xsl:variable name="type" select="$schema/xs:complexType[@name = $ref/@type]"/>
            
            <xsl:choose>
                <xsl:when test="exists($type/@abstract)">
                    <xsl:for-each select="$schema/xs:element[@substitutionGroup = $name]">
                        <xsl:variable name="specific-name" select="./@name"/><!--
                            <xsl:apply-templates select="copy-of($element)">
                            <xsl:with-param name="specific-name" select="$specific-name"/>
                            </xsl:apply-templates>-->
                        <xsl:element name="xs:element">
                            <xsl:attribute name="ref" select="$specific-name"></xsl:attribute>
                            <xsl:attribute name="minOccurs" select="$element/@minOccurs"></xsl:attribute>
                            <xsl:attribute name="maxOccurs" select="$element/@maxOccurs"></xsl:attribute>
                        </xsl:element>
                    </xsl:for-each>
                </xsl:when>
                <xsl:otherwise>
                    <xsl:sequence select="$element"/>
                </xsl:otherwise>
            </xsl:choose>
        </xsl:for-each>
    </xsl:function>
    
    <xsl:template match="xs:element">
        <xsl:param name="specific-name"/>
        
        <xsl:attribute name="ref" select="$specific-name"/>
        
        <xsl:sequence select="."></xsl:sequence>
    </xsl:template>
    
</xsl:stylesheet>
