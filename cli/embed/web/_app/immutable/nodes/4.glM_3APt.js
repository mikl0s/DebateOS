const __vite__mapDeps=(i,m=__vite__mapDeps,d=(m.f||(m.f=["../chunks/DmMN-uJ2.js","../chunks/BlsMDL08.js","../chunks/CXOfwQjV.js","../chunks/Df4EiJ7t.js"])))=>i.map(i=>d[i]);
import{i as C,p as G,r as Z,s as ae,_ as pe,a as Me,b as Ne}from"../chunks/BlsMDL08.js";import{b as Oe,a as f,t as x,c as re,n as Ie}from"../chunks/BpVn7OHi.js";import{h as Q,A as ue,e as ve,b as De,E as Ee,a as We,aH as Te,z as Le,x as Ve,C as ge,y as he,K as je,G as Be,F as Fe,d as He,s as w,c as v,r as c,t as N,aI as Ye,p as $,aG as E,a6 as oe,f as Ge,i as n,g as ee,aE as U,ag as D,am as Ke,$ as qe}from"../chunks/CXOfwQjV.js";import{s as W}from"../chunks/DljT5etN.js";import{h as Je}from"../chunks/D9A97xFR.js";import{d as xe,g as Qe,w as Ue,o as Xe}from"../chunks/Df4EiJ7t.js";import{b as fe}from"../chunks/BsB74Sq8.js";import{e as q,i as X}from"../chunks/B-jJNYo9.js";import{a as be,b as ne,s as K}from"../chunks/CYhT1aF9.js";import{f as Ze,d as ie}from"../chunks/BAO5_qdH.js";import{s as $e}from"../chunks/Do_BejOy.js";function ea(a,e,r,l,p,z){let m=Q;Q&&ue();var h,u,s=null;Q&&ve.nodeType===1&&(s=ve,ue());var t=Q?ve:a,y;De(()=>{const b=e()||null;var M=Te;b!==h&&(y&&(b===null?Be(y,()=>{y=null,u=null}):b===u?Fe(y):He(y)),b&&b!==u&&(y=We(()=>{if(s=Q?s:document.createElementNS(M,b),Oe(s,s),l){Q&&Ze(b)&&s.append(document.createComment(""));var A=Q?Le(s):s.appendChild(Ve());Q&&(A===null?ge(!1):he(A)),l(s,A)}je.nodes_end=s,t.before(s)})),h=b,h&&(u=h))},Ee),m&&(ge(!0),he(t))}const aa=!1,ta=!1,gt=Object.freeze(Object.defineProperty({__proto__:null,prerender:ta,ssr:aa},Symbol.toStringTag,{value:"Module"}));function ra(){const a={foundation:"arch",panes:[],hardware:{predicates:[]}},{subscribe:e,set:r,update:l}=Ue(a);let p=0;return{subscribe:e,setFoundation(z){l(m=>({...m,foundation:z}))},addPane(z,m,h,u){const s=`pane-${++p}`;return l(t=>({...t,panes:[...t.panes,{id:s,pointId:z,pointName:m,curator:u,opinions:h}]})),s},removePane(z){l(m=>({...m,panes:m.panes.filter(h=>h.id!==z)}))},reorderPanes(z){l(m=>({...m,panes:z}))},updateHardware(z){l(m=>({...m,hardware:z}))},resetDebate(){p=0,r(a)},snapshot(){return Qe({subscribe:e})}}}const Y=ra();xe(Y,a=>a.panes.flatMap(e=>e.opinions));xe(Y,a=>a.panes.length);function oa(a){const{state:e,icon:r,label:l}=na(a.rule);return{state:e,icon:r,label:l,text:a.text,opinionsInvolved:a.opinions_involved??[],kept:a.kept??[],dropped:a.dropped??[],patchOffered:a.patch_offered??"",hasPatch:!!(a.patch_offered&&a.patch_offered.length>0),trustWarning:a.trust_warning??"",alternativeSuggestion:a.alternative_suggestion??""}}function na(a){switch(a){case"rule2":return{state:"hard",icon:"AlertTriangle",label:"Hard conflict"};case"cycle":return{state:"hard",icon:"AlertTriangle",label:"Circular dependency"};case"rule1":case"rule3":return{state:"warn",icon:"Info",label:"Will be dropped"};case"sysctl-collision":return{state:"warn",icon:"Info",label:"Sysctl collision"};case"hardware-skip":return{state:"hardware",icon:"Cpu",label:"Hardware mismatch"};case"hardware-apply":return{state:"info",icon:"Cpu",label:"Hardware applied"};case"no-conflict":return{state:"compat",icon:"CheckCircle2",label:"Compatible"};case"rule4":return{state:"patch",icon:"Puzzle",label:"Patch applied"};case"ordering":return{state:"info",icon:"Info",label:"Ordering applied"};default:return{state:"info",icon:"Info",label:a}}}var ia=x(`<button style="
				min-height: var(--min-height-touch);
				padding: 0 var(--spacing-md);
				border: 1px solid var(--color-border-subtle);
				border-radius: 6px;
				background: transparent;
				color: var(--color-text-secondary);
				font-size: var(--font-size-label);
				cursor: pointer;
				display: flex;
				align-items: center;
				gap: var(--spacing-xs);
			">Swap Foundation</button>`),la=x(`<div style="
		height: var(--height-foundation-bar);
		background-color: var(--color-surface-card);
		border-bottom: 1px solid var(--color-border-subtle);
		display: flex;
		align-items: center;
		padding: 0 var(--spacing-lg);
		gap: var(--spacing-md);
		flex-shrink: 0;
	"><span style="
			font-size: var(--font-size-label);
			color: var(--color-text-secondary);
			text-transform: uppercase;
			letter-spacing: 0.06em;
		">Foundation</span> <span style="
			font-size: var(--font-size-heading);
			font-weight: 600;
			color: var(--color-text-primary);
			text-transform: capitalize;
		"> </span> <div style="flex: 1;"></div> <!></div>`);function sa(a,e){var r=la(),l=w(v(r),2),p=v(l,!0);c(l);var z=w(l,4);{var m=h=>{var u=ia();u.__click=function(...s){var t;(t=e.onSwap)==null||t.apply(this,s)},f(h,u)};C(z,h=>{e.onSwap&&h(m)})}c(r),N(()=>W(p,e.foundation)),f(a,r)}ie(["click"]);/**
 * @file
 * @license @lucide/svelte v1.18.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const ca={xmlns:"http://www.w3.org/2000/svg",width:24,height:24,viewBox:"0 0 24 24",fill:"none",stroke:"currentColor","stroke-width":2,"stroke-linecap":"round","stroke-linejoin":"round"};/**
 * @file
 * @license @lucide/svelte v1.18.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const da=a=>{for(const e in a)if(e.startsWith("aria-")||e==="role"||e==="title")return!0;return!1};/**
 * @file
 * @license @lucide/svelte v1.18.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const va=Symbol("lucide-context"),fa=()=>Ye(va);var pa=Ie("<svg><!><!></svg>");function te(a,e){$(e,!0);const r=fa()??{},l=G(e,"color",19,()=>r.color??"currentColor"),p=G(e,"size",19,()=>r.size??24),z=G(e,"strokeWidth",19,()=>r.strokeWidth??2),m=G(e,"absoluteStrokeWidth",19,()=>r.absoluteStrokeWidth??!1),h=G(e,"iconNode",19,()=>[]),u=Z(e,["$$slots","$$events","$$legacy","name","color","size","strokeWidth","absoluteStrokeWidth","iconNode","children"]),s=E(()=>m()?Number(z())*24/Number(p()):z());var t=pa();let y;var b=v(t);q(b,17,h,X,(A,V)=>{let L=()=>n(V)[0],T=()=>n(V)[1];var O=re(),j=oe(O);ea(j,L,!0,(k,d)=>{let o;N(()=>o=be(k,o,{...T()}))}),f(A,O)});var M=w(b);$e(M,()=>e.children??Ge),c(t),N(A=>y=be(t,y,{...ca,...A,...u,width:p(),height:p(),stroke:l(),"stroke-width":n(s),class:["lucide-icon lucide",r.class,e.name&&`lucide-${e.name}`,e.class]}),[()=>!e.children&&!da(u)&&{"aria-hidden":"true"}]),f(a,t),ee()}function _e(a,e){let r=Z(e,["$$slots","$$events","$$legacy"]);te(a,ae({name:"circle-check"},()=>r,{iconNode:[["circle",{cx:"12",cy:"12",r:"10"}],["path",{d:"m9 12 2 2 4-4"}]]}))}function we(a,e){let r=Z(e,["$$slots","$$events","$$legacy"]);te(a,ae({name:"circle-x"},()=>r,{iconNode:[["circle",{cx:"12",cy:"12",r:"10"}],["path",{d:"m15 9-6 6"}],["path",{d:"m9 9 6 6"}]]}))}function ua(a,e){let r=Z(e,["$$slots","$$events","$$legacy"]);te(a,ae({name:"cpu"},()=>r,{iconNode:[["path",{d:"M12 20v2"}],["path",{d:"M12 2v2"}],["path",{d:"M17 20v2"}],["path",{d:"M17 2v2"}],["path",{d:"M2 12h2"}],["path",{d:"M2 17h2"}],["path",{d:"M2 7h2"}],["path",{d:"M20 12h2"}],["path",{d:"M20 17h2"}],["path",{d:"M20 7h2"}],["path",{d:"M7 20v2"}],["path",{d:"M7 2v2"}],["rect",{x:"4",y:"4",width:"16",height:"16",rx:"2"}],["rect",{x:"8",y:"8",width:"8",height:"8",rx:"1"}]]}))}function ga(a,e){let r=Z(e,["$$slots","$$events","$$legacy"]);te(a,ae({name:"info"},()=>r,{iconNode:[["circle",{cx:"12",cy:"12",r:"10"}],["path",{d:"M12 16v-4"}],["path",{d:"M12 8h.01"}]]}))}function ze(a,e){let r=Z(e,["$$slots","$$events","$$legacy"]);te(a,ae({name:"puzzle"},()=>r,{iconNode:[["path",{d:"M15.39 4.39a1 1 0 0 0 1.68-.474 2.5 2.5 0 1 1 3.014 3.015 1 1 0 0 0-.474 1.68l1.683 1.682a2.414 2.414 0 0 1 0 3.414L19.61 15.39a1 1 0 0 1-1.68-.474 2.5 2.5 0 1 0-3.014 3.015 1 1 0 0 1 .474 1.68l-1.683 1.682a2.414 2.414 0 0 1-3.414 0L8.61 19.61a1 1 0 0 0-1.68.474 2.5 2.5 0 1 1-3.014-3.015 1 1 0 0 0 .474-1.68l-1.683-1.682a2.414 2.414 0 0 1 0-3.414L4.39 8.61a1 1 0 0 1 1.68.474 2.5 2.5 0 1 0 3.014-3.015 1 1 0 0 1-.474-1.68l1.683-1.682a2.414 2.414 0 0 1 3.414 0z"}]]}))}function ke(a,e){let r=Z(e,["$$slots","$$events","$$legacy"]);te(a,ae({name:"triangle-alert"},()=>r,{iconNode:[["path",{d:"m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3"}],["path",{d:"M12 9v4"}],["path",{d:"M12 17h.01"}]]}))}var ha=x("<div><!> <span> </span></div>");function me(a,e){const r={hard:"var(--color-conflict-hard)",warn:"var(--color-conflict-warn)",hardware:"var(--color-conflict-warn)",compat:"var(--color-conflict-compat)",patch:"var(--color-accent-brand)",info:"var(--color-text-secondary)"},l=E(()=>r[e.state]??r.info);var p=ha(),z=v(p);{var m=t=>{ke(t,{size:16,"aria-hidden":"true","data-icon":"AlertTriangle"})},h=(t,y)=>{{var b=A=>{_e(A,{size:16,"aria-hidden":"true","data-icon":"CheckCircle2"})},M=(A,V)=>{{var L=O=>{ua(O,{size:16,"aria-hidden":"true","data-icon":"Cpu"})},T=(O,j)=>{{var k=o=>{ze(o,{size:16,"aria-hidden":"true","data-icon":"Puzzle"})},d=o=>{ga(o,{size:16,"aria-hidden":"true","data-icon":"Info"})};C(O,o=>{e.icon==="Puzzle"?o(k):o(d,!1)},j)}};C(A,O=>{e.icon==="Cpu"?O(L):O(T,!1)},V)}};C(t,A=>{e.icon==="CheckCircle2"?A(b):A(M,!1)},y)}};C(z,t=>{e.icon==="AlertTriangle"?t(m):t(h,!1)})}var u=w(z,2),s=v(u,!0);c(u),c(p),N(()=>{ne(p,`
		display: inline-flex;
		align-items: center;
		gap: var(--spacing-xs);
		color: ${n(l)??""};
		font-size: var(--font-size-label);
		font-weight: 600;
	`),W(s,e.label)}),f(a,p)}var ba=x(`<p style="
					font-size: var(--font-size-body);
					color: var(--color-text-secondary);
					margin: 0;
					line-height: var(--line-height-body);
				"> </p>`),ma=(a,e,r)=>{var l;return(l=e.onDrop)==null?void 0:l.call(e,n(r))},ya=x(`<button class="conflict-action-btn" style="
								min-height: var(--min-height-touch);
								padding: 0 var(--spacing-md);
								border: 1px solid var(--color-destructive);
								border-radius: 6px;
								background: transparent;
								color: var(--color-destructive);
								font-size: var(--font-size-label);
								cursor: pointer;
								display: flex;
								align-items: center;
							"> </button>`),xa=(a,e)=>{var r;return(r=e.onApplyPatch)==null?void 0:r.call(e,e.view.patchOffered)},_a=x(`<button class="conflict-action-btn" style="
							min-height: var(--min-height-touch);
							padding: 0 var(--spacing-md);
							border: 1px solid var(--color-accent-brand);
							border-radius: 6px;
							background: var(--color-accent-brand);
							color: #ffffff;
							font-size: var(--font-size-label);
							cursor: pointer;
							display: flex;
							align-items: center;
						">Apply Patch</button>`),wa=x('<div style="display: flex; gap: var(--spacing-sm); flex-wrap: wrap;"><!> <!></div>'),za=x('<div role="status"><div style="display: flex; align-items: center; gap: var(--spacing-sm); flex-wrap: wrap;"><!> <!></div> <!> <!></div>');function ka(a,e){$(e,!0);let r=G(e,"opinionsLabel",3,"");const l={hard:{bg:"var(--color-conflict-hard)",bgOpacity:.18,border:"var(--color-conflict-hard)",borderStyle:"solid"},warn:{bg:"var(--color-conflict-warn)",bgOpacity:.14,border:"var(--color-conflict-warn)",borderStyle:"solid"},hardware:{bg:"var(--color-conflict-warn)",bgOpacity:.14,border:"var(--color-conflict-warn)",borderStyle:"dashed"},compat:{bg:"var(--color-conflict-compat)",bgOpacity:.12,border:"var(--color-conflict-compat)",borderStyle:"dashed"},patch:{bg:"var(--color-accent-brand)",bgOpacity:.1,border:"var(--color-accent-brand)",borderStyle:"solid"},info:{bg:"transparent",bgOpacity:0,border:"var(--color-border-subtle)",borderStyle:"solid"}},p=E(()=>l[e.view.state]??l.info),z=E(()=>`${e.view.label}${r()?": "+r():""}`),m=E(()=>e.view.state==="hard"?`rgba(239, 68, 68, ${n(p).bgOpacity})`:e.view.state==="warn"||e.view.state==="hardware"?`rgba(245, 158, 11, ${n(p).bgOpacity})`:e.view.state==="compat"?`rgba(34, 197, 94, ${n(p).bgOpacity})`:e.view.state==="patch"?`rgba(99, 102, 241, ${n(p).bgOpacity})`:"transparent"),h=E(()=>e.view.state==="hard"?"239,68,68":e.view.state==="warn"||e.view.state==="hardware"?"245,158,11":e.view.state==="compat"?"34,197,94":e.view.state==="patch"?"99,102,241":"");var u=re(),s=oe(u);{var t=y=>{var b=za(),M=v(b),A=v(M);me(A,{get state(){return e.view.state},get icon(){return e.view.icon},get label(){return e.view.label}});var V=w(A,2);{var L=d=>{me(d,{state:"patch",icon:"Puzzle",label:"Patch available"})};C(V,d=>{e.view.hasPatch&&d(L)})}c(M);var T=w(M,2);{var O=d=>{var o=ba(),g=v(o,!0);c(o),N(()=>W(g,e.view.text)),f(d,o)};C(T,d=>{e.view.text&&d(O)})}var j=w(T,2);{var k=d=>{var o=wa(),g=v(o);q(g,17,()=>e.view.dropped,X,(_,R)=>{var S=re(),I=oe(S);{var B=J=>{var F=ya();F.__click=[ma,e,R];var le=v(F);c(F),N(()=>W(le,`Drop '${n(R)??""}'`)),f(J,F)};C(I,J=>{e.onDrop&&J(B)})}f(_,S)});var i=w(g,2);{var P=_=>{var R=_a();R.__click=[xa,e],f(_,R)};C(i,_=>{e.view.hasPatch&&e.onApplyPatch&&_(P)})}c(o),f(d,o)};C(j,d=>{(e.view.state==="hard"||e.view.state==="warn")&&d(k)})}c(b),N(()=>{K(b,"aria-label",n(z)),K(b,"data-conflict-state",e.view.state),K(b,"data-conflict-bg-rgb",n(h)),ne(b,`
			position: relative;
			border-radius: 6px;
			padding: var(--spacing-sm) var(--spacing-md);
			background-color: ${n(m)??""};
			border: 2px ${n(p).borderStyle??""} ${n(p).border??""};
			display: flex;
			flex-direction: column;
			gap: var(--spacing-sm);
		`)}),f(y,b)};C(s,y=>{e.view.state!=="info"&&y(t)})}f(a,u),ee()}ie(["click"]);function ye(a,e,r){var l;n(e)?((l=r.onRemove)==null||l.call(r,r.pane.id),D(e,!1)):D(e,!0)}function Pa(a,e){D(e,!1)}var Ca=x(`<span style="
						font-size: var(--font-size-label);
						color: var(--color-text-secondary);
					"> </span>`),Aa=x(`<button class="conflict-action-btn" style="
					min-height: var(--min-height-touch);
					min-width: var(--min-height-touch);
					display: flex;
					align-items: center;
					justify-content: center;
					border: none;
					background: transparent;
					color: var(--color-text-secondary);
					cursor: pointer;
					border-radius: 4px;
					padding: 0 var(--spacing-sm);
				"><!></button>`),Ra=x(`<div style="display: flex; align-items: center; gap: var(--spacing-xs);"><span style="font-size: var(--font-size-label); color: var(--color-destructive);">Remove pane?</span> <button style="
						min-height: var(--min-height-touch);
						padding: 0 var(--spacing-sm);
						border: 1px solid var(--color-border-subtle);
						border-radius: 4px;
						background: transparent;
						font-size: var(--font-size-label);
						cursor: pointer;
						color: var(--color-text-secondary);
					">Cancel</button> <button class="conflict-action-btn" style="
						min-height: var(--min-height-touch);
						padding: 0 var(--spacing-sm);
						border: 1px solid var(--color-destructive);
						border-radius: 4px;
						background: var(--color-destructive);
						color: #ffffff;
						font-size: var(--font-size-label);
						cursor: pointer;
					">Remove</button></div>`),Sa=x(`<div style="
					padding: var(--spacing-xs) var(--spacing-sm);
					border-radius: 4px;
					background-color: var(--color-surface-base);
					border: 1px solid var(--color-border-subtle);
					display: flex;
					align-items: center;
					gap: var(--spacing-sm);
				"><span style="font-size: var(--font-size-label); color: var(--color-text-primary); flex: 1;"> </span> <span style="
						font-size: var(--font-size-label);
						color: var(--color-text-secondary);
						text-transform: capitalize;
					"> </span></div>`),Ma=x('<div style="padding: 0 var(--spacing-md) var(--spacing-md);"></div>'),Na=x(`<div role="region"><div class="pane-header" style="
			min-height: var(--min-height-touch);
			padding: 0 var(--spacing-md);
			display: flex;
			align-items: center;
			gap: var(--spacing-sm);
			background-color: var(--color-surface-card);
			border-bottom: 1px solid var(--color-border-subtle);
		"><div style="flex: 1; display: flex; flex-direction: column; gap: 2px;"><span style="
					font-size: var(--font-size-heading);
					font-weight: 600;
					color: var(--color-text-primary);
					line-height: var(--line-height-heading);
				"> </span> <!></div> <!></div> <div style="padding: var(--spacing-sm) var(--spacing-md); display: flex; flex-direction: column; gap: var(--spacing-xs);"></div> <!></div>`);function Oa(a,e){$(e,!0);let r=G(e,"conflictViews",19,()=>[]),l=G(e,"isActive",3,!1),p=U(!1);const z=E(()=>new Set(e.pane.opinions.map(k=>k.id))),m=E(()=>r().filter(k=>k.state!=="info"&&k.opinionsInvolved.some(d=>n(z).has(d))));var h=Na(),u=v(h),s=v(u),t=v(s),y=v(t,!0);c(t);var b=w(t,2);{var M=k=>{var d=Ca(),o=v(d);c(d),N(()=>W(o,`by ${e.pane.curator??""}`)),f(k,d)};C(b,k=>{e.pane.curator&&k(M)})}c(s);var A=w(s,2);{var V=k=>{var d=Aa();d.__click=[ye,p,e];var o=v(d);we(o,{size:18,"aria-hidden":"true"}),c(d),N(()=>K(d,"aria-label",`Remove ${e.pane.pointName??""} pane`)),f(k,d)},L=k=>{var d=Ra(),o=w(v(d),2);o.__click=[Pa,p];var g=w(o,2);g.__click=[ye,p,e],c(d),f(k,d)};C(A,k=>{n(p)?k(L,!1):k(V)})}c(u);var T=w(u,2);q(T,21,()=>e.pane.opinions,X,(k,d)=>{var o=Sa(),g=v(o),i=v(g,!0);c(g);var P=w(g,2),_=v(P,!0);c(P),c(o),N(()=>{W(i,n(d).name),W(_,n(d).status)}),f(k,o)}),c(T);var O=w(T,2);{var j=k=>{var d=Ma();q(d,21,()=>n(m),X,(o,g)=>{const i=E(()=>n(g).opinionsInvolved.join(" vs "));ka(o,{get view(){return n(g)},get opinionsLabel(){return n(i)},get onDrop(){return e.onDrop},get onApplyPatch(){return e.onApplyPatch}})}),c(d),f(k,d)};C(O,k=>{n(m).length>0&&k(j)})}c(h),N(()=>{K(h,"aria-label",`${e.pane.pointName??""} pane`),K(h,"data-pane-id",e.pane.id),K(h,"data-pane-active",l()?"true":"false"),ne(h,`
		background-color: var(--color-surface-pane);
		border: 1px solid ${(l()?"var(--color-accent-brand)":"var(--color-border-subtle)")??""};
		border-radius: 8px;
		overflow: hidden;
		${(l()?"box-shadow: 0 0 0 2px var(--color-accent-brand);":"")??""}
	`),W(y,e.pane.pointName)}),f(a,h),ee()}ie(["click"]);var Ia=x(`<div style="
					display: flex;
					flex-direction: column;
					align-items: center;
					justify-content: center;
					text-align: center;
					padding: var(--spacing-3xl) var(--spacing-lg);
					gap: var(--spacing-md);
				"><h2 style="
						font-size: var(--font-size-display);
						font-weight: 600;
						color: var(--color-text-primary);
						margin: 0;
					">Your debate has no points yet.</h2> <p style="
						font-size: var(--font-size-body);
						color: var(--color-text-secondary);
						max-width: 480px;
						margin: 0;
						line-height: var(--line-height-body);
					">Browse the registry to find points from curators you trust, or add your own opinions
					directly. Your foundation is chosen — time to take a stand.</p> <a style="
						display: inline-flex;
						align-items: center;
						justify-content: center;
						min-height: var(--min-height-touch);
						padding: 0 var(--spacing-xl);
						background-color: var(--color-accent-brand);
						color: #ffffff;
						font-size: var(--font-size-body);
						font-weight: 600;
						text-decoration: none;
						border-radius: 6px;
					">Browse Points</a></div>`),Da=x(`<div style="
		display: flex;
		flex-direction: column;
		height: 100%;
		overflow: hidden;
	"><!> <div style="
			flex: 1;
			overflow-y: auto;
			padding: var(--spacing-md);
			display: flex;
			flex-direction: column;
			gap: var(--spacing-md);
		"><!></div></div>`);function Ea(a,e){$(e,!0);let r=G(e,"conflictViews",19,()=>[]);var l=Da(),p=v(l);sa(p,{get foundation(){return e.foundation},get onSwap(){return e.onSwapFoundation}});var z=w(p,2),m=v(z);{var h=s=>{var t=Ia(),y=w(v(t),4);K(y,"href",`${fe??""}/browse/`),c(t),f(s,t)},u=s=>{var t=re(),y=oe(t);q(y,17,()=>e.panes,b=>b.id,(b,M)=>{const A=E(()=>n(M).id===e.activePaneId);Oa(b,{get pane(){return n(M)},get conflictViews(){return r()},get onRemove(){return e.onRemovePane},get onDrop(){return e.onDropOpinion},get onApplyPatch(){return e.onApplyPatch},get isActive(){return n(A)}})}),f(s,t)};C(m,s=>{e.panes.length===0?s(h):s(u,!1)})}c(z),c(l),f(a,l),ee()}var Wa=x(`<p class="explanation-text" style="
				font-size: var(--font-size-body);
				line-height: var(--line-height-body);
				color: var(--color-text-primary);
				margin: 0;
			"> </p>`),Ta=x(`<span style="
						padding: 2px var(--spacing-sm);
						background-color: var(--color-surface-base);
						border: 1px solid var(--color-border-subtle);
						border-radius: 4px;
						font-size: var(--font-size-label);
						color: var(--color-text-secondary);
					"> </span>`),La=x('<div style="display: flex; flex-wrap: wrap; gap: var(--spacing-xs);"></div>'),Va=x('<span style="font-size: var(--font-size-label); color: var(--color-text-secondary);"> </span>'),ja=x('<div style="display: flex; align-items: center; gap: var(--spacing-sm); flex-wrap: wrap;"><!> <span style="font-size: var(--font-size-label); color: var(--color-conflict-compat); font-weight: 600;">Kept:</span> <!></div>'),Ba=x('<span style="font-size: var(--font-size-label); color: var(--color-text-secondary);"> </span>'),Fa=x('<div style="display: flex; align-items: center; gap: var(--spacing-sm); flex-wrap: wrap;"><!> <span style="font-size: var(--font-size-label); color: var(--color-conflict-hard); font-weight: 600;">Dropped:</span> <!></div>'),Ha=(a,e)=>{var r;return(r=e.onApplyPatch)==null?void 0:r.call(e,e.view.patchOffered)},Ya=x(`<button style="
						min-height: var(--min-height-touch);
						padding: 0 var(--spacing-md);
						border: 1px solid var(--color-accent-brand);
						border-radius: 6px;
						background: var(--color-accent-brand);
						color: #ffffff;
						font-size: var(--font-size-label);
						cursor: pointer;
						display: flex;
						align-items: center;
					">Apply Resolution</button>`),Ga=x('<div style="display: flex; align-items: center; gap: var(--spacing-sm); flex-wrap: wrap;"><!> <span style="font-size: var(--font-size-label); color: var(--color-accent-brand);"> </span> <!></div>'),Ka=x(`<div style="
				display: flex;
				align-items: flex-start;
				gap: var(--spacing-sm);
				padding: var(--spacing-sm) var(--spacing-md);
				background-color: rgba(245, 158, 11, 0.12);
				border: 1px solid var(--color-conflict-warn);
				border-radius: 6px;
			"><!> <p style="margin: 0; font-size: var(--font-size-label); color: var(--color-text-primary);"> </p></div>`),qa=x(`<section style="
		background-color: var(--color-surface-card);
		border: 1px solid var(--color-border-subtle);
		border-radius: 8px;
		padding: var(--spacing-md);
		display: flex;
		flex-direction: column;
		gap: var(--spacing-sm);
	"><div style="display: flex; align-items: center; gap: var(--spacing-sm);"><span> </span></div> <!> <!> <!> <!> <!> <!></section>`);function Ja(a,e){$(e,!0);const r={hard:"Rule 2",warn:"Rule 1",hardware:"Hardware",compat:"Compatible",patch:"Rule 4",info:"Info"},l=E(()=>r[e.view.state]??e.view.state),p={hard:"var(--color-conflict-hard)",warn:"var(--color-conflict-warn)",hardware:"var(--color-conflict-warn)",compat:"var(--color-conflict-compat)",patch:"var(--color-accent-brand)",info:"var(--color-text-secondary)"},z=E(()=>p[e.view.state]??p.info);var m=qa(),h=v(m),u=v(h),s=v(u,!0);c(u),c(h);var t=w(h,2);{var y=o=>{var g=Wa(),i=v(g,!0);c(g),N(()=>W(i,e.view.text)),f(o,g)};C(t,o=>{e.view.text&&o(y)})}var b=w(t,2);{var M=o=>{var g=La();q(g,21,()=>e.view.opinionsInvolved,X,(i,P)=>{var _=Ta(),R=v(_,!0);c(_),N(()=>W(R,n(P))),f(i,_)}),c(g),f(o,g)};C(b,o=>{e.view.opinionsInvolved.length>0&&o(M)})}var A=w(b,2);{var V=o=>{var g=ja(),i=v(g);_e(i,{size:16,color:"var(--color-conflict-compat)","aria-hidden":"true"});var P=w(i,4);q(P,17,()=>e.view.kept,X,(_,R)=>{var S=Va(),I=v(S,!0);c(S),N(()=>W(I,n(R))),f(_,S)}),c(g),f(o,g)};C(A,o=>{e.view.kept.length>0&&o(V)})}var L=w(A,2);{var T=o=>{var g=Fa(),i=v(g);we(i,{size:16,color:"var(--color-conflict-hard)","aria-hidden":"true"});var P=w(i,4);q(P,17,()=>e.view.dropped,X,(_,R)=>{var S=Ba(),I=v(S,!0);c(S),N(()=>W(I,n(R))),f(_,S)}),c(g),f(o,g)};C(L,o=>{e.view.dropped.length>0&&o(T)})}var O=w(L,2);{var j=o=>{var g=Ga(),i=v(g);ze(i,{size:16,color:"var(--color-accent-brand)","aria-hidden":"true"});var P=w(i,2),_=v(P);c(P);var R=w(P,2);{var S=I=>{var B=Ya();B.__click=[Ha,e],f(I,B)};C(R,I=>{e.onApplyPatch&&I(S)})}c(g),N(()=>W(_,`Patch available: ${e.view.patchOffered??""}`)),f(o,g)};C(O,o=>{e.view.hasPatch&&o(j)})}var k=w(O,2);{var d=o=>{var g=Ka(),i=v(g);ke(i,{size:16,color:"var(--color-conflict-warn)",style:"flex-shrink: 0; margin-top: 2px;","aria-hidden":"true"});var P=w(i,2),_=v(P,!0);c(P),c(g),N(()=>W(_,e.view.trustWarning)),f(o,g)};C(k,o=>{e.view.trustWarning&&o(d)})}c(m),N(()=>{ne(u,`
				display: inline-block;
				padding: 2px var(--spacing-sm);
				border-radius: 999px;
				background-color: ${n(z)??""};
				color: #ffffff;
				font-size: var(--font-size-label);
				font-weight: 600;
			`),W(s,n(l))}),f(a,m),ee()}ie(["click"]);var Qa=x('<p style="color: var(--color-text-secondary); font-size: var(--font-size-body); margin: 0;">No conflicts detected. Add more points to your debate.</p>'),Ua=x(`<div style="
				border: 1px solid var(--color-conflict-compat);
				border-radius: 8px;
				padding: var(--spacing-md);
				background-color: rgba(34, 197, 94, 0.08);
				display: flex;
				flex-direction: column;
				gap: var(--spacing-sm);
			"><h3 style="
					font-size: var(--font-size-heading);
					font-weight: 600;
					color: var(--color-conflict-compat);
					margin: 0;
				">Your speech is ready.</h3> <p style="font-size: var(--font-size-body); color: var(--color-text-secondary); margin: 0;">No conflicts remain. Run the command below to build your installer, or download the
				resolved speech YAML to build later.</p> <a style="
					display: inline-flex;
					align-items: center;
					justify-content: center;
					min-height: var(--min-height-touch);
					padding: 0 var(--spacing-xl);
					background-color: var(--color-accent-brand);
					color: #ffffff;
					font-size: var(--font-size-body);
					font-weight: 600;
					text-decoration: none;
					border-radius: 6px;
					align-self: flex-start;
				">Proceed to Build</a></div>`),Xa=x(`<aside style="
		display: flex;
		flex-direction: column;
		gap: var(--spacing-md);
		padding: var(--spacing-md);
		background-color: var(--color-surface-card);
		border-left: 1px solid var(--color-border-subtle);
		overflow-y: auto;
		min-width: 280px;
		max-width: 380px;
	"><h2 style="
			font-size: var(--font-size-heading);
			font-weight: 600;
			color: var(--color-text-primary);
			margin: 0;
		">Resolution Panel</h2> <!> <!> <!></aside>`);function Za(a,e){$(e,!0);const r=E(()=>e.views.filter(t=>t.state!=="info")),l=E(()=>e.views.some(t=>t.state==="hard"));var p=Xa(),z=w(v(p),2);{var m=t=>{var y=Qa();f(t,y)};C(z,t=>{n(r).length===0&&e.applied.length===0&&t(m)})}var h=w(z,2);q(h,17,()=>n(r),t=>t.state+t.text.slice(0,20),(t,y)=>{Ja(t,{get view(){return n(y)},get onApplyPatch(){return e.onApplyPatch}})});var u=w(h,2);{var s=t=>{var y=Ua(),b=w(v(y),4);K(b,"href",`${fe??""}/export/`),c(y),f(t,y)};C(u,t=>{e.isReady&&e.applied.length>0&&!n(l)&&t(s)})}c(p),f(a,p),ee()}var $a=x(`<div role="status" aria-live="polite" data-wasm-ready="false" style="
			flex: 1;
			display: flex;
			align-items: center;
			justify-content: center;
			color: var(--color-text-secondary);
			font-size: var(--font-size-body);
		">Loading the resolver…</div>`),et=x(`<div role="alert" data-wasm-ready="error" style="
			flex: 1;
			display: flex;
			align-items: center;
			justify-content: center;
			color: var(--color-conflict-hard);
			font-size: var(--font-size-body);
			text-align: center;
			max-width: 480px;
			margin: 0 auto;
		">The resolver failed to load. Refresh to try again. Your debate is saved in this browser session.</div>`),at=x(`<div role="status" aria-live="polite" style="
						padding: var(--spacing-xs) var(--spacing-md);
						font-size: var(--font-size-label);
						color: var(--color-text-secondary);
						background-color: var(--color-surface-card);
						border-top: 1px solid var(--color-border-subtle);
					">Resolving…</div>`),tt=x(`<div role="status" aria-live="polite" style="
						padding: var(--spacing-xs) var(--spacing-md);
						font-size: var(--font-size-label);
						color: var(--color-conflict-compat);
						background-color: var(--color-surface-card);
						border-top: 1px solid var(--color-border-subtle);
					"> </div>`),rt=x(`<div data-wasm-ready="true" style="
			flex: 1;
			display: flex;
			overflow: hidden;
		"><div style="flex: 1; min-width: 0; display: flex; flex-direction: column; overflow: hidden;"><!> <!></div> <!></div>`);function ht(a,e){$(e,!0);const[r,l]=Me(),p=()=>Ne(Y,"$debate",r);let z=U(!1),m=U(!1),h=U(""),u=U(null),s=U(Ke([])),t=U(null),y=U(!1),b=null;Xe(async()=>{try{const{loadDebateosWasm:i}=await pe(async()=>{const{loadDebateosWasm:_}=await import("../chunks/DmMN-uJ2.js");return{loadDebateosWasm:_}},__vite__mapDeps([0,1,2,3]),import.meta.url);return await i(fe),D(z,!0),Y.subscribe(()=>{M()})}catch(i){D(m,!0),D(h,i instanceof Error?i.message:String(i),!0),console.error("WASM load error:",i)}});function M(){b&&clearTimeout(b),b=setTimeout(()=>A(),150)}async function A(){const{debateosResolve:i,buildResolveInput:P}=await pe(async()=>{const{debateosResolve:R,buildResolveInput:S}=await import("../chunks/DmMN-uJ2.js");return{debateosResolve:R,buildResolveInput:S}},__vite__mapDeps([0,1,2,3]),import.meta.url),_=Y.snapshot();if(_.panes.length===0){D(u,null),D(s,[],!0),D(t,null);return}D(y,!0),D(t,null);try{const R={schema:1,foundation:_.foundation,points:_.panes.map(F=>({id:F.pointId})),opinions:void 0,hardware:void 0},S=_.panes.flatMap(F=>F.opinions),I=P(R,S,_.hardware),{resolved:B,error:J}=i(I);D(u,B,!0),D(t,J??null,!0),D(s,(B.explanations??[]).map(oa),!0)}catch(R){D(t,R instanceof Error?R.message:String(R),!0),D(s,[],!0)}finally{D(y,!1)}}const V=E(()=>{var i;return n(u)!==null&&(((i=n(u).applied)==null?void 0:i.length)??0)>0&&!n(s).some(P=>P.state==="hard")}),L=E(()=>{var i,P;return((P=(i=n(u))==null?void 0:i.applied)==null?void 0:P.length)??0});function T(i){Y.removePane(i)}function O(i){const _=Y.snapshot().panes.find(R=>R.opinions.some(S=>S.id===i));_&&Y.removePane(_.id)}function j(i){M()}typeof window<"u"&&(window.debateAddTestPane=(i,P,_)=>{Y.addPane(i,P,_)},window.debateGetResolved=()=>n(u));var k=re();Je(i=>{qe.title="Debate — DebateOS"});var d=oe(k);{var o=i=>{var P=$a();f(i,P)},g=(i,P)=>{{var _=S=>{var I=et();f(S,I)},R=S=>{var I=rt(),B=v(I),J=v(B);Ea(J,{get foundation(){return p().foundation},get panes(){return p().panes},get conflictViews(){return n(s)},onRemovePane:T,onDropOpinion:O,onApplyPatch:j});var F=w(J,2);{var le=H=>{var se=at();f(H,se)},Pe=(H,se)=>{{var Re=ce=>{var de=tt(),Se=v(de);c(de),N(()=>W(Se,`Resolved · ${n(L)??""} opinion${(n(L)!==1?"s":"")??""} applied`)),f(ce,de)};C(H,ce=>{n(u)&&ce(Re)},se)}};C(F,H=>{n(y)?H(le):H(Pe,!1)})}c(B);var Ce=w(B,2);const Ae=E(()=>{var H;return((H=n(u))==null?void 0:H.applied)??[]});Za(Ce,{get views(){return n(s)},get applied(){return n(Ae)},get isReady(){return n(V)},onApplyPatch:j}),c(I),f(S,I)};C(i,S=>{n(m)?S(_):S(R,!1)},P)}};C(d,i=>{!n(z)&&!n(m)?i(o):i(g,!1)})}f(a,k),ee(),l()}export{ht as component,gt as universal};
